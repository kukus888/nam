// Extension for sending JSON with forms via hx-ext="submitjson"
htmx.defineExtension('submitjson', {
    onEvent: function (name, evt) {
        if (name === "htmx:configRequest") {
            evt.detail.headers['Content-Type'] = "application/json"
        }
    },
    encodeParameters: function (xhr, parameters, elt) {
        xhr.overrideMimeType('text/json') // override default mime type

        // Filter out empty string values
        const filteredParameters = {};
        for (const [key, value] of Object.entries(parameters)) {
            if (value !== "") {
                filteredParameters[key] = value;
            }
        }

        let json = JSON.stringify(filteredParameters);

        // Remove quotes from numbers (preserve previous behavior)
        const regex = /"(-|)([0-9]+(?:\.[0-9]+)?)"/g;
        json = json.replace(regex, '$1$2');
        // Remove quotes from booleans
        json = json.replace(/"true"/g, 'true').replace(/"false"/g, 'false');

        return json;
    }
})

function handleRestApiResponse(event, redirect) {
    if(event.detail.successful) {
        htmx.ajax('GET',redirect, {target: '#items-container'})
    } else {
        showErrorMessage(event.detail.xhr.responseText)
    }
}