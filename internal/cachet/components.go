package cachet

import (
	"context"
	"encoding/json"
	"net/url"
	"sort"
	"strconv"
)

// PageName is the human name of the status page. Cachet's public API does not
// expose the page title, so we hard-code ArvanCloud's.
const PageName = "ArvanCloud"

// Component status values (Cachet). Higher is worse.
const (
	CompOperational = 1
	CompPerformance = 2
	CompPartial     = 3
	CompMajor       = 4
)

// Service is one node of the status tree. Component groups are the top level
// (e.g. a region) and components are their children. Children recurse, though
// Cachet only nests one level deep in practice.
type Service struct {
	ID       int
	Name     string
	Status   Status
	Children []Service
}

// Summary is the assembled component tree plus page metadata, mirroring what a
// status page renders.
type Summary struct {
	PageName string
	Services []Service
}

// ServicesByID walks the tree once and returns a map keyed by component ID.
// Only leaf components are included: group IDs and component IDs come from
// separate Cachet sequences and overlap, and incidents only ever reference
// component IDs, so mapping groups too would let a group clobber a same-ID
// component.
func (s *Summary) ServicesByID() map[int]Service {
	out := map[int]Service{}
	var walk func([]Service)
	walk = func(xs []Service) {
		for _, x := range xs {
			if len(x.Children) == 0 {
				out[x.ID] = x
			}
			walk(x.Children)
		}
	}
	walk(s.Services)
	return out
}

// componentAttr is the attributes object of a `components` resource.
type componentAttr struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Order  *int   `json:"order"`
	Status Status `json:"status"`
}

// groupAttr is the attributes object of a `componentGroups` resource.
type groupAttr struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Order int    `json:"order"`
}

// FetchSummary retrieves all component groups and components and assembles them
// into a region → service tree. Components without a group are collected under
// a synthetic "Other" group so nothing is dropped.
func (c *Client) FetchSummary(ctx context.Context) (*Summary, error) {
	groups, err := c.fetchGroups(ctx)
	if err != nil {
		return nil, err
	}
	return c.buildSummary(ctx, groups)
}

func (c *Client) fetchGroups(ctx context.Context) ([]groupAttr, error) {
	var doc document
	if err := c.getJSON(ctx, "component-groups", url.Values{"per_page": {"200"}}, &doc); err != nil {
		return nil, err
	}
	var rs []resource
	if err := json.Unmarshal(doc.Data, &rs); err != nil {
		return nil, err
	}
	groups := make([]groupAttr, 0, len(rs))
	for _, r := range rs {
		var g groupAttr
		if err := json.Unmarshal(r.Attributes, &g); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, nil
}

func (c *Client) buildSummary(ctx context.Context, groups []groupAttr) (*Summary, error) {
	var doc document
	q := url.Values{"include": {"group"}, "per_page": {"200"}}
	if err := c.getJSON(ctx, "components", q, &doc); err != nil {
		return nil, err
	}
	var rs []resource
	if err := json.Unmarshal(doc.Data, &rs); err != nil {
		return nil, err
	}

	// Bucket components by group ID; 0 == ungrouped.
	byGroup := map[int][]Service{}
	for _, r := range rs {
		var a componentAttr
		if err := json.Unmarshal(r.Attributes, &a); err != nil {
			return nil, err
		}
		svc := Service{ID: a.ID, Name: a.Name, Status: a.Status}

		gid := 0
		if rel, ok := r.Relationships["group"]; ok {
			if ref, ok := rel.one(); ok {
				if id, err := strconv.Atoi(ref.ID); err == nil {
					gid = id
				}
			}
		}
		byGroup[gid] = append(byGroup[gid], svc)
	}

	sort.Slice(groups, func(i, j int) bool { return groups[i].Order < groups[j].Order })

	services := make([]Service, 0, len(groups)+1)
	for _, g := range groups {
		kids := byGroup[g.ID]
		delete(byGroup, g.ID)
		if len(kids) == 0 {
			continue
		}
		sortServices(kids)
		services = append(services, Service{
			ID:       g.ID,
			Name:     g.Name,
			Status:   worstStatus(kids),
			Children: kids,
		})
	}

	// Any remaining buckets are components whose group is missing/ungrouped.
	var orphans []Service
	for _, kids := range byGroup {
		orphans = append(orphans, kids...)
	}
	if len(orphans) > 0 {
		sortServices(orphans)
		services = append(services, Service{
			ID:       0,
			Name:     "Other",
			Status:   worstStatus(orphans),
			Children: orphans,
		})
	}

	return &Summary{PageName: PageName, Services: services}, nil
}

func sortServices(xs []Service) {
	sort.Slice(xs, func(i, j int) bool { return xs[i].Name < xs[j].Name })
}

// worstStatus returns the highest (worst) component status among the children,
// so a group's bullet reflects its most degraded member.
func worstStatus(xs []Service) Status {
	worst := Status{Human: "Operational", Value: CompOperational}
	for _, x := range xs {
		if x.Status.Value > worst.Value {
			worst = x.Status
		}
	}
	return worst
}
