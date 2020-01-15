package members

import (
	"fmt"
	"html/template"
	"sort"
)

func (s *server) index() {
	email, err := s.googleEmail()
	if err != nil {
		s.redirect(absURL("/login"))
		return
	}

	var data struct {
		Name       string
		Email      string
		Total      int
		Attendance []struct {
			Month        int
			Community    int // sloppy boolean
			DirectAction int // sloppy boolean
			Events       []struct {
				Date         string // "YYYY-MM-DD"
				Name         string
				Community    int // sloppy boolean
				DirectAction int // sloppy boolean
			}
		}
	}

	const q = `
select json_object(
  'Name', a.name,
  'Email', a.email,
  'Attendance', (
    select coalesce(json_arrayagg(months), '[]') from (
      select json_object(
        'Month', e.month,
        'Community', max(e.community),
        'DirectAction', max(e.direct_action),
        'Events', json_arrayagg(json_object(
          'Date', e.date,
          'Name', e.name,
          'Community', e.community,
          'DirectAction', e.direct_action
        ))
      ) as months
      from (
        select id, name, date,
               extract(year_month from date) as month,
               event_type in ('Circle', 'Community', 'Training') as community,
               event_type in ('Action', 'Campaign Action', 'Frontline Surveillance', 'Outreach', 'Sanctuary') as direct_action
        from events
      ) e
      join event_attendance ea
        on (e.id = ea.event_id)
      where ea.activist_id = a.id
      group by e.month
    )
  )
)
from activists a
where a.email = ?
  and not hidden
`

	if err := s.queryJSON(&data, q, email); err != nil {
		// TODO(mdempsky): Error might be because the user
		// doesn't have an activist record. We should suggest
		// they try logging in as someone else.
		s.error(err)
		return
	}

	// Workaround MySQL 5.7 limitations:

	// Manually sort in descending order by date, as MySQL doesn't
	// allow control of json_arrayagg()'s aggregation order.
	sort.Slice(data.Attendance, func(i, j int) bool { return data.Attendance[i].Month > data.Attendance[j].Month })
	for k := range data.Attendance {
		events := data.Attendance[k].Events
		sort.Slice(events, func(i, j int) bool { return events[i].Date > events[j].Date })

		// Manually calculate total events, as MySQL doesn't
		// support LATERAL joins until 8.0.
		data.Total += len(events)
	}

	s.render(indexTmpl, &data)
}

var indexTmpl = template.Must(template.New("index").Funcs(template.FuncMap{
	"monthfmt": func(n int) string { return fmt.Sprintf("%d-%02d", n/100, n%100) },
}).Parse(`
<!doctype html>
<html>
<head>
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<link href="https://fonts.googleapis.com/css?family=Source+Sans+Pro&display=swap" rel="stylesheet">

<style>
body {
  font-family: 'Source Sans Pro', sans-serif;
}

.wrap {
  max-width: 40em;
  margin-left: auto;
  margin-right: auto;
}

table {
  padding-top: 2em;
  border-spacing: 0;
  font-variant-numeric: tabular-nums;
}

tr.month {
  background-color: #ddd;
  font-weight: bold;
}

tr.month td {
  text-align: center;
}

tr.mpi {
  background-color: #beb;
}

td {
  padding: 0.375em;
}

td:nth-child(3) {
  white-space: nowrap;
}

.green { background-color: #beb; }
.gray { background-color: #ddd; }
</style>
</head>

<body>
<div class="wrap">

<p>Hello, <b>{{.Name}}</b> ({{.Email}})! (Not you? <a href="login?force">Click here</a> to login as someone else.)</p>

<p>Below is a list of the <b>{{.Total}}</b> events you've attended with DxE SF.</p>

<p>🏙️ indicates the event is a "community" event;<br>
📣 indicates a "direct action" event.</p>

<p>A <b class="green">green</b> bar indicates you met the MPI requirements for that month;<br>
a <b class="gray">gray</b> bar indicates you did not.</p>

<p>Questions or feedback? Email <a href="mailto:tech@dxe.io">tech@dxe.io</a>.</p>

<table>
{{range .Attendance}}
<tr class="month {{if and .Community .DirectAction}}mpi{{end}}">
  <td>{{if .Community}}🏙️{{end}}</td>
  <td>{{if .DirectAction}}📣{{end}}</td>
  <td colspan=2>{{monthfmt .Month}}</td>
</tr>
{{range .Events}}
<tr>
  <td>{{if .Community}}🏙️{{end}}</td>
  <td>{{if .DirectAction}}📣{{end}}</td>
  <td>{{.Date}}</td>
  <td>{{.Name}}</td>
</tr>
{{end}}
{{end}}
</table>

</div>
</body>
</html>
`))
