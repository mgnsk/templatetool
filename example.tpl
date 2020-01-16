{{define "Template1"}}
    {{- .MyVar1 | Title}}
{{end}}
    
{{define "Template2"}}
    {{- .MyVar2 | Title}}
{{end}}

{{define "MapTemplate"}}
{{- .Header}}
{{/* For very precise whitespace control, print statements should be used */}}
    {{- range $sectionTitle, $items := .Lists}}
<section>
    {{$sectionTitle}}:
    {{- range $item := $items}}
    <p>{{$item}}</p>
    {{- end}}
</section>
{{end}}
{{end}}
