// Extension for sending JSON with forms via hx-ext="submitjson"
htmx.defineExtension('submitjson', {
    onEvent: function (name, evt) {
        if (name === "htmx:configRequest") {
            evt.detail.headers['Content-Type'] = "application/json"
            evt.detail.headers['X-API-Key'] = 'sjk_xxx' // TODO: API sec stuff
        }
    },
    encodeParameters: function (xhr, parameters, elt) {
        xhr.overrideMimeType('text/json') // override default mime type
        json = JSON.stringify(parameters)
        const regex = /"(-|)([0-9]+(?:\.[0-9]+)?)"/g 
        json = json.replace(regex, '$1$2')
        return json
    }
})

function handleRestApiResponse(event, redirect) {
    if(event.detail.successful) {
        htmx.ajax('GET',redirect, {target: '#items-container'})
    } else {
        showErrorMessage(event.detail.xhr.responseText)
    }
}