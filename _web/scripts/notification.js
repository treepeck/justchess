// TODO: allow multiple notifications.

const notification = document.createElement("div")
notification.classList.add("notification")

const closeBtn = document.createElement("button")
closeBtn.textContent = "X"

export default function showNotification(message) {
	notification.textContent = message
	document.body.appendChild(notification)
	notification.appendChild(closeBtn)

	closeBtn.addEventListener("click", hideNotification)

	closeBtn.focus()
}

function hideNotification() {
	closeBtn.removeEventListener(closeBtn, hideNotification)
	document.body.removeChild(notification)
}
