/**
 * Displays a popup message.
 * Popup is a modal dialog box with some content.
 * @param {string} msg
 */
export function popup(msg) {
	const html = `
		<dialog id="popup" class="wrap">
			<p>${msg}</p>
			<div class="dialog-buttons">
				<button commandfor="popup" command="close">Close</button>
			</div>
		</dialog>`

	const template = document.createElement("template")
	template.innerHTML = html

	const container = document.getElementById("main")
	container.appendChild(template.content)

	const dialog = document.getElementById("popup")
	dialog.showModal()
}
