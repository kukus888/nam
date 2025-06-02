# healthcheck components

Nomenclature: `health_<component>_<size>`
component - appDef = Application definition
size - tiny, small, medium, large

```html
<div hx-get="/htmx/health/application/instance?live_reload=true&size=small&id={{ .ID }}"
    hx-swap="outerHTML" hx-trigger="load"
    hx-on::after-request="if(!event.detail.successful) showErrorMessage(event.detail.xhr.responseText)">
    <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-gray-100 text-gray-800">
        Loading...
    </span>
</div>
```