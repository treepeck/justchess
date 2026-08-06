// Engine difficulty slider for engineConfig dialog.
const slider = document.getElementById("difficultySlider")
const demo = document.getElementById("sliderDemo")
slider.addEventListener("input", (e) => {
	switch (e.target.value) {
	case "1":
		demo.textContent = "Easy"
		slider.style.setProperty("--bg", `url("/assets/images/easy.svg")`)
		break
	case "2":
		demo.textContent = "Medium"
		slider.style.setProperty("--bg", `url("/assets/images/medium.svg")`)
		break
	case "3":
		demo.textContent = "Hard"
		slider.style.setProperty("--bg", `url("/assets/images/hard.svg")`)
		break
	case "4":
		demo.textContent = "Insane"
		slider.style.setProperty("--bg", `url("/assets/images/insane.svg")`)
		break
	case "5":
		demo.textContent = "Impossible"
		slider.style.setProperty("--bg", `url("/assets/images/impossible.svg")`)
		break
	}
})