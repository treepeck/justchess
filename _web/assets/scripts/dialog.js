// Close dialog on by clicking its backgrop.
for (const dialog of document.getElementsByTagName("dialog")) {
	dialog.onclick = (e) => {
		const rect = dialog.getBoundingClientRect()

		const isInsideDialog =
			rect.left <= e.clientX &&
			e.clientX <= rect.left + rect.width &&
			rect.top <= e.clientY &&
			e.clientY <= rect.top + rect.height

		if (!isInsideDialog) dialog.close()
	}
}
