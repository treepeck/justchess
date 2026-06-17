// Format "Member since"
const l = new Intl.DateTimeFormat(undefined, {
	dateStyle: "long",
})
for (const time of document.getElementsByTagName("time")) {
	time.textContent = l.format(new Date(time.textContent))
}
