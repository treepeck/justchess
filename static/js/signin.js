// Regular expressions to validate user input.
const emailEx = /^[a-zA-Z0-9._]+@[a-zA-Z0-9._]+\.[a-zA-Z0-9._]+$/
const pwdEx   = /^[a-zA-Z0-9!@#$%^&*()_+-/.<>]{5,71}$/

function emailChange(e) {
    removePopup()

    const container = document.getElementById("email-container")

    if (!emailEx.test(e.target.value)) {
        if (document.getElementById("email-error") === null) {
            const error = document.createElement("p")
            error.id = "email-error"
            error.textContent = "Please enter a valid email address."
            container.appendChild(error)
        }
    } else if (document.getElementById("email-error") !== null) {
        container.removeChild(document.getElementById("email-error"))
    }
}

function passwordChange(e) {
    removePopup()

    const container = document.getElementById("password-container")

    if (!pwdEx.test(e.target.value)) {
        if (document.getElementById("password-error") === null) {
            const error = document.createElement("p")
            error.id = "password-error"
            error.textContent = "Password must contain 5 to 71 characters (english, numbers, special symbols)."
            container.appendChild(error)
        }
    } else if (document.getElementById("password-error") !== null) {
        container.removeChild(document.getElementById("password-error"))
    }
}

function visibilityChange(e) {
    if (e.target.classList.contains("show")) {
        document.getElementById("password").type = "text"
        e.target.classList.remove("show")
        e.target.classList.add("hide")
    } else {
        document.getElementById("password").type = "password"
        e.target.classList.remove("hide")
        e.target.classList.add("show")
    }
}

function removePopup(e) {
    const popup = document.getElementById("popup")
    if (popup) {
        document.getElementsByTagName("main")[0].removeChild(popup)
    }
}

function createPopup(message) {
    // Delete the previous popup if it exists.
    let popup = document.getElementById("popup")
    if (popup !== null) popup.remove()

    // Create popup element.
    popup = document.createElement("div")
    popup.id = "popup"

    // Appent popup content.
    popup.textContent = message

    // Show the popup.
    document.getElementsByTagName("main")[0].append(popup)
}

async function submitForm(e) {
    e.preventDefault()

    const form = document.getElementById("form")
    const formData = new FormData(form)

    if (!emailEx.test(formData.get("email")) ||
        !pwdEx.test(formData.get("password"))) {
        createPopup("Please fill all fields.")
        return 
    }

    try {
        const res = await fetch("http://localhost:3502/auth/signin", {
            method: "POST",
            headers: {
                "Content-Type": "application/x-www-form-urlencoded",
            },
            body: new URLSearchParams(formData).toString()
        })

        if (res.ok) {
            window.location.replace("/index.html")
        } else {
            const text = await res.text()

            createPopup(text)
        }
    } catch (e) {
        createPopup(e)
    }
}

document.getElementById("email").addEventListener("change", emailChange)
document.getElementById("password").addEventListener("change", passwordChange)
document.getElementById("visibility-icon").addEventListener("click", visibilityChange)
document.getElementById("submit").addEventListener("click", async (e) => { await submitForm(e) })