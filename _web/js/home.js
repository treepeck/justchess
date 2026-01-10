
// Show and hide help window.
helpText.addEventListener("click", () => {
	helpWindow.classList.toggle("show")

	// Focus close button.
	closeBtn.focus()
})

closeBtn.addEventListener("click", () => {
	helpWindow.classList.toggle("show")
})