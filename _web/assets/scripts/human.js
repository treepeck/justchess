const slider = document.getElementById("slider")
const demo = document.getElementById("sliderDemo")
slider.addEventListener("input", (e) => {
	switch (e.target.value) {
		case "0":
			demo.textContent = "Ultra bullet"
			slider.style.setProperty("--slider-thumb-image", `url("/assets/images/ultra-bullet.svg")`)
			break
		case "1":
			demo.textContent = "Bullet"
			slider.style.setProperty("--slider-thumb-image", `url("/assets/images/bullet.svg")`)
			break
		case "2":
			demo.textContent = "Fast Blitz"
			slider.style.setProperty("--slider-thumb-image", `url("/assets/images/fast-blitz.svg")`)
			break
		case "3":
			demo.textContent = "Fast Blitz with Increment"
			slider.style.setProperty("--slider-thumb-image", `url("/assets/images/fast-blitz-with-increment.svg")`)
			break
		case "4":
			demo.textContent = "Blitz"
			slider.style.setProperty(
				"--slider-thumb-image",
				`url("/assets/images/blitz.svg")`,
			)
			break
		case "5":
			demo.textContent = "Blitz with Increment"
			slider.style.setProperty(
				"--slider-thumb-image",
				`url("/assets/images/blitz-with-increment.svg")`,
			)
			break
		case "6":
			demo.textContent = "Rapid"
			slider.style.setProperty("--slider-thumb-image", `url("/assets/images/rapid.svg")`)
			break
		case "7":
			demo.textContent = "Rapid with Increment"
			slider.style.setProperty("--slider-thumb-image", `url("/assets/images/rapid-with-increment.svg")`)
			break
		case "8":
			demo.textContent = "Time to think"
			slider.style.setProperty("--slider-thumb-image", `url("/assets/images/time-to-think.svg")`)
			break
	}
})
// Set default slider thumb image.
demo.textContent = "Fast Blitz with Increment"
	slider.style.setProperty(
	"--slider-thumb-image",
	`url("/assets/images/fast-blitz-with-increment.svg")`,
)
