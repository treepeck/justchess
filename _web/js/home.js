for (let i = 1; i <= 9; i++) {
	const cell = document.getElementById(`cell${i}`)
	cell.addEventListener("click", () => {
		// Redirect the user to the queue page.
		window.location.href = `http://localhost:3502/queue?id=${i}`
	})
}

// Show and hide help window.
helpText.addEventListener("click", () => {
	helpWindow.classList.toggle("show")

	// Focus close button.
	closeHelp.focus()
})

closeHelp.addEventListener("click", () => {
	helpWindow.classList.toggle("show")
})
