package template

import (
	"cmp"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"
	"text/template"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func Run(args []string) {
	flags := struct {
		json     string
		jsonFile string
		tmpl     string
		tmplFile string
	}{}

	fs := flag.NewFlagSet("template", flag.ExitOnError)
	fs.StringVar(&flags.json, "json", "", "JSON string")
	fs.StringVar(&flags.jsonFile, "json-file", "", "JSON file")
	fs.StringVar(&flags.tmpl, "tmpl", "", "Template string")
	fs.StringVar(&flags.tmplFile, "tmpl-file", "", "Template file")

	if err := fs.Parse(args[1:]); err != nil {
		fs.Usage()
		os.Exit(1)
	}

	// Validate that at least one of the JSON input flags is provided.
	if flags.json == "" && flags.jsonFile == "" {
		fmt.Fprintln(os.Stderr, "An input JSON string must be provided via either -json or -json-file flag")
		fs.Usage()
		os.Exit(1)
	}

	// Validate that at least one of the template input flags is provided.
	if flags.tmpl == "" && flags.tmplFile == "" {
		fmt.Fprintln(os.Stderr, "A template string must be provided via either -tmpl or -tmpl-file flag")
		fs.Usage()
		os.Exit(1)
	}

	notif, err := parseJSON(flags.json, flags.jsonFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	template, err := parseTemplate(flags.tmpl, flags.tmplFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = template.Execute(os.Stdout, notif)
	if err != nil && err != io.EOF {
		fmt.Fprintln(os.Stderr, "Error executing template:", err)
		os.Exit(1)
	}

	fmt.Println()
}

func parseJSON(content, file string) (*Notification, error) {
	if content == "" {
		b, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("Error reading JSON file: %w", err)
		}

		content = string(b)
	}

	notif := &Notification{}
	if err := json.Unmarshal([]byte(content), notif); err != nil {
		return nil, fmt.Errorf("Error parsing JSON: %w", err)
	}

	return notif, nil
}

func parseTemplate(content, file string) (*template.Template, error) {
	if content == "" {
		b, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("Error reading template file: %w", err)
		}

		content = string(b)
	}

	return template.New("tpl").Funcs(funcs).Parse(content)
}

type Notification struct {
	OrgID             int    `json:"orgId"`
	Receiver          string `json:"receiver"`
	Status            string `json:"status"`
	State             string `json:"state"`
	Alerts            Alerts `json:"alerts"`
	GroupLabels       Map    `json:"groupLabels"`
	CommonLabels      Map    `json:"commonLabels"`
	CommonAnnotations Map    `json:"commonAnnotations"`
	ExternalURL       string `json:"externalURL"`
	GroupKey          string `json:"groupKey"`
	TruncatedAlerts   int    `json:"truncatedAlerts"`
}

type Alerts []Alert

func (a Alerts) filter(status string) Alerts {
	result := make(Alerts, 0)
	for _, alert := range a {
		if alert.Status == status {
			result = append(result, alert)
		}
	}

	return result
}

func (a Alerts) Firing() Alerts {
	return a.filter("firing")
}

func (a Alerts) Resolved() Alerts {
	return a.filter("resolved")
}

type Alert struct {
	OrgID        int       `json:"orgId"`
	RuleUID      string    `json:"ruleUID"`
	Status       string    `json:"status"`
	Labels       Map       `json:"labels"`
	Annotations  Map       `json:"annotations"`
	StartsAt     time.Time `json:"startsAt"`
	EndsAt       time.Time `json:"endsAt"`
	GeneratorURL string    `json:"generatorURL"`
	DashboardURL string    `json:"dashboardURL"`
	PanelURL     string    `json:"panelURL"`
	SilenceURL   string    `json:"silenceURL"`
	Values       Map       `json:"values"`
	ValueString  string    `json:"valueString"`
	Fingerprint  string    `json:"fingerprint"`
}

type Map map[string]string

func (m Map) SortedPairs() Pairs {
	pairs := make(Pairs, 0, len(m))
	for name, value := range m {
		pairs = append(pairs, Pair{Name: name, Value: value})
	}

	slices.SortFunc(pairs, func(a, b Pair) int {
		return cmp.Compare(a.Name, b.Name)
	})

	return pairs
}

func (m Map) Remove(keys []string) Map {
	clone := make(Map, len(m))
	for k, v := range m {
		clone[k] = v
	}

	for _, k := range keys {
		delete(clone, k)
	}

	return clone
}

func (m Map) Names() []string {
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}

	return names
}

func (m Map) Values() []string {
	values := make([]string, 0, len(m))
	for _, value := range m {
		values = append(values, value)
	}

	return values
}

type Pairs []Pair

func (p Pairs) Names() []string {
	names := make([]string, 0, len(p))
	for _, pair := range p {
		names = append(names, pair.Name)
	}

	return names
}

func (p Pairs) Values() []string {
	values := make([]string, 0, len(p))
	for _, pair := range p {
		values = append(values, pair.Value)
	}

	return values
}

type Pair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

var funcs = template.FuncMap{
	"title":     cases.Title(language.English).String,
	"toUpper":   strings.ToUpper,
	"toLower":   strings.ToLower,
	"trimSpace": strings.TrimSpace,
	"match":     regexp.MatchString,

	"reReplaceAll": func(pattern, replacement, text string) (string, error) {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return "", err
		}

		return re.ReplaceAllString(text, replacement), nil
	},

	"join": func(sep string, s []string) string {
		return strings.Join(s, sep)
	},

	"safeHtml": template.HTMLEscapeString,

	"stringSlice": func(s ...string) []string {
		return s
	},

	"date": func(layout string, t time.Time) string {
		return t.Format(layout)
	},

	"tz": func(timezone string, t time.Time) (time.Time, error) {
		loc, err := time.LoadLocation(timezone)
		if err != nil {
			return t, err
		}

		return t.In(loc), nil
	},

	// Namespaced functions
	"coll": func() CollFuncs { return CollFuncs{} },
	"data": func() DataFuncs { return DataFuncs{} },
	"time": func() TimeFuncs { return TimeFuncs{} },
}

type CollFuncs struct{}

func (CollFuncs) Dict(pairs ...any) Map {
	m := make(Map, len(pairs)/2)

	for i := 0; i < len(pairs); i += 2 {
		name := fmt.Sprintf("%v", pairs[i])

		var value string
		if i+1 < len(pairs) {
			value = fmt.Sprintf("%v", pairs[i+1])
		}

		m[name] = value
	}

	return m
}

func (CollFuncs) Slice(values ...any) []any {
	return values
}

func (CollFuncs) Append(value any, list []any) []any {
	return append(list, value)
}

type DataFuncs struct{}

func (DataFuncs) JSON(s string) (any, error) {
	var obj any
	if err := json.Unmarshal([]byte(s), &obj); err != nil {
		return nil, fmt.Errorf("data.JSON: error parsing JSON: %w", err)
	}

	return obj, nil
}

func (DataFuncs) ToJSON(obj any) (string, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("data.ToJSON: error serializing JSON: %w", err)
	}

	return string(b), nil
}

func (DataFuncs) ToJSONPretty(indent string, obj any) (string, error) {
	b, err := json.MarshalIndent(obj, "", indent)
	if err != nil {
		return "", fmt.Errorf("data.ToJSONPretty: error serializing JSON: %w", err)
	}

	return string(b), nil
}

type TimeFuncs struct{}

func (TimeFuncs) Now() time.Time {
	return time.Now()
}
