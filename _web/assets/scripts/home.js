import "/assets/scripts/widget/engineSlider.js"
import { createEngine } from "/assets/scripts/api/api.js"

for (let i = 0; i < 9; i++) {
	const btn = document.getElementById(`control${i}`)
	btn.addEventListener("click", () => {
		window.location.href = `/queue/${i}`
	})
}

document.getElementById("playEngine").addEventListener("click", async () => {
	const difficulty = parseInt(
		document.getElementById("difficultySlider").value,
	)
	if (!difficulty) throw new Error("engine difficulty not set")

	await createEngine(difficulty)
})
