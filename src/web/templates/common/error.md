Error Notification Component Documentation
How to Use This Component
Include the component in your main layout or on pages where you need error notifications:

Html
Insert code
```html
{{ template "template/components/error-notification" . }}
```
Trigger the error notification from anywhere in your JavaScript code:

Tsx
Insert code
```js
// Display a simple error message
showErrorMessage("Something went wrong. Please try again.");

// Display an error with JSON data
showErrorMessage({
  error: "Failed to save data",
  trace: "Error details: Connection refused at line 42"
});

// Pass a JSON string (common from API responses)
showErrorMessage('{"error": "Authentication failed", "trace": {"code": 401, "details": "Token expired"}}');
```
Integrate with HTMX error handling:

Html
Insert code
```html
<form hx-post="/api/endpoint" 
      hx-on::after-request="if(!event.detail.successful) showErrorMessage(event.detail.xhr.responseText)">
  <!-- form content -->
</form>
```
Programmatically hide the notification:

Tsx
Insert code

// Hide the error notification when needed
hideErrorNotification();
Features
Persistent Display: Error notifications remain visible until explicitly dismissed by the user
Technical Details: Expandable section for detailed error information (stack traces, etc.)
JSON Support: Automatically parses and displays JSON error responses
Object Handling: Properly formats object error messages and traces
Responsive Design: Adapts to different screen sizes with a maximum width of 48rem (768px)
Slide-in Animation: Smooth entrance animation from the top of the screen
Manual Dismiss: Can be dismissed by clicking the close button or outside interactive elements
Overflow Protection: Long error messages are contained with scrolling
Accessibility: Includes proper contrast, focus states, and semantic structure
Error Format Support
The component handles various error formats:

Simple String: "Something went wrong"
JSON String: '{"error": "Database connection failed"}'
Error Object: {error: "File not found", trace: "Error at line 27"}
Complex Object: {error: "API Error", trace: {code: 500, details: {...}}}
Technical Details
When an error contains trace information (via trace or stack properties), a "Show technical details" button appears. Clicking this button reveals the detailed error information, which is particularly useful for developers during debugging.

The component intelligently formats trace data:

String traces are displayed as-is
Object traces are pretty-printed as JSON
Styling
The notification uses Tailwind CSS with:

Red color scheme for error indication
Responsive width (max-width: 3xl / 48rem)
Shadow and border styling for emphasis
Proper spacing and typography for readability
Overflow handling for long content
Behavior
No Auto-dismiss: Notifications stay visible until manually dismissed
Click Handling: Clicking on technical details or interactive elements won't dismiss the notification
Close Button: Dedicated close button in the top-right corner
Animation: Smooth transition when showing/hiding
Implementation Notes
The component uses vanilla JavaScript without dependencies on hyperscript or other libraries, making it compatible with any project that includes Tailwind CSS.