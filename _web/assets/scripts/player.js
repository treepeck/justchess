import "/assets/scripts/time/format.js"

const humanPane = document.getElementById("humanPane")
const enginePane = document.getElementById("enginePane")

document.getElementById("tabLink1").onclick = () => {
	// Do not handle cliks if this pane is already selected.
	if (humanPane.classList.contains("active")) {
		return
	}

	humanPane.classList.add("active")
	enginePane.classList.remove("active")
}
document.getElementById("tabLink2").onclick = () => {
	// Do not handle cliks if this pane is already selected.
	if (enginePane.classList.contains("active")) {
		return
	}

	enginePane.classList.add("active")
	humanPane.classList.remove("active")
}
