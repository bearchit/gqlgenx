package graphiql

import (
	"html/template"
	"net/http"
)

var page = template.Must(template.New("graphiql").Parse(`
<html>
  <head>
    <title>{{.title}}</title>
    <link href="https://unpkg.com/graphiql/graphiql.min.css" rel="stylesheet" />
  </head>
  <body style="margin: 0;">
    <div id="graphiql" style="height: 100vh;"></div>

    <script
      crossorigin
      src="https://unpkg.com/react/umd/react.production.min.js"
    ></script>
    <script
      crossorigin
      src="https://unpkg.com/react-dom/umd/react-dom.production.min.js"
    ></script>
    <script
      crossorigin
      src="https://unpkg.com/graphiql/graphiql.min.js"
    ></script>

    <script>
      const graphQLFetcher = graphQLParams =>
        fetch('{{.endpoint}}', {
          method: 'post',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(graphQLParams),
        })
          .then(response => response.json())
          .catch(() => response.text());
      ReactDOM.render(
        React.createElement(GraphiQL, { fetcher: graphQLFetcher }),
        document.getElementById('graphiql'),
      );
    </script>
  </body>
</html>
`))

func Handler(title string, endpoint string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		err := page.Execute(w, map[string]string{
			"title":    title,
			"endpoint": endpoint,
		})
		if err != nil {
			panic(err)
		}
	}
}
