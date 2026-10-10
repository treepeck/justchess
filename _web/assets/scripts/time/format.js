// format.js takes all timestamp elements and formats them according to the
// player's time zone.

const intl = new Intl.DateTimeFormat("en-GB", {
	month: "short",
	day: "2-digit",
	year: "numeric",
})

for (const ts of document.querySelectorAll(".timestamp")) {
	ts.textContent = intl.format(new Date(ts.textContent))
}
