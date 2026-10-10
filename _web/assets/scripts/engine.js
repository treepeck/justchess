const slider = document.getElementById("slider")
const demo = document.getElementById("sliderDemo")
slider.addEventListener("input", (e) => {
	switch (e.target.value) {
		case "1":
			demo.textContent = "Easy"
			slider.style.setProperty(
				"--slider-thumb-image",
				`url("/assets/images/easy.svg")`,
			)
			break
		case "2":
			demo.textContent = "Medium"
			slider.style.setProperty(
				"--slider-thumb-image",
				`url("/assets/images/medium.svg")`,
			)
			break
		case "3":
			demo.textContent = "Hard"
			slider.style.setProperty(
				"--slider-thumb-image",
				`url("/assets/images/hard.svg")`,
			)
			break
		case "4":
			demo.textContent = "Insane"
			slider.style.setProperty(
				"--slider-thumb-image",
				`url("/assets/images/insane.svg")`,
			)
			break
		case "5":
			demo.textContent = "Impossible"
			slider.style.setProperty(
				"--slider-thumb-image",
				`url("/assets/images/impossible.svg")`,
			)
			break
	}
})
// Set default slider thumb image.
demo.textContent = "Medium"
slider.style.setProperty(
	"--slider-thumb-image",
	`url("/assets/images/medium.svg")`,
)
