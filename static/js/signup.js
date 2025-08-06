// Regular expressions to validate user input.
const nameEx  = /^[a-zA-Z0-9]{2,60}$/
const emailEx = /^[a-zA-Z0-9._]+@[a-zA-Z0-9._]+\.[a-zA-Z0-9._]+$/
const pwdEx   = /^[a-zA-Z0-9!@#$%^&*()_+-/.<>]{5,71}$/

function emailChange(e) {
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

function nameChange(e) {
    const container = document.getElementById("name-container")

    if (!nameEx.test(e.target.value)) {
        if (document.getElementById("name-error") === null) {
            const error = document.createElement("p")
            error.id = "name-error"
            error.textContent = "Must contain 2 to 60 characters (english letters, numbers)."
            container.appendChild(error)
        }
    } else if (document.getElementById("name-error") !== null) {
        container.removeChild(document.getElementById("name-error"))
    }
}

function passwordChange(e) {
    const container = document.getElementById("password-container")

    if (!pwdEx.test(e.target.value)) {
        if (document.getElementById("password-error") === null) {
            const error = document.createElement("p")
            error.id = "password-error"
            error.textContent = "Must contain 5 to 71 characters (english, numbers, special symbols)."
            container.appendChild(error)
        }
    } else if (document.getElementById("password-error") !== null) {
        container.removeChild(document.getElementById("password-error"))
    }
}

async function submitForm(e) {
    e.preventDefault()
    console.log(document.getElementByTagName("form"))
    const formData = new FormData(document.getElementsByTagName("form"))
    if (!nameEx.test(formData.name) || !emailEx.test(formData.email) ||
        !pwdEx.test(formData.password)) {
        return
    }

    try {
        const res = await fetch("http://localhost:3502/auth/signup", {
            method: "POST",
            body: formData
        })

        if (res.ok) {
            window.location.replace("/singin")
        } else {
            const text = await res.text()

            const error = document.createElement("p")
            error.textContent = text
            document.getElementByTagName("form").appendChild(error)
        }
    } catch (e) {
        console.error(e)
    }
}

document.getElementById("name").addEventListener("change", nameChange)
document.getElementById("email").addEventListener("change", emailChange)
document.getElementById("password").addEventListener("change", passwordChange)
document.getElementById("submit").addEventListener("click", async (e) => { await submitForm(e) })