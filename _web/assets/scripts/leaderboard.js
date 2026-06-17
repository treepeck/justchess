// Formats dates to a localized strings.
// Intl.DateTimeFormat is reused to format all dates on the document.
// This is way more efficient than repetitive calls of Date.toLocaleString()

// Initialize with empty constructor to use the client's browser settings.
const l = new Intl.DateTimeFormat(undefined, {
	dateStyle: "long",
})
l.timeStyle
for (const time of document.getElementsByTagName("time")) {
	console.log(time.textContent)
	time.textContent = l.format(new Date(time.textContent))
}
