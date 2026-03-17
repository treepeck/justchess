const nameEx = /^[a-zA-Z0-9]{2,60}$/i
const emailEx = /^[a-zA-Z0-9._]+@[a-zA-Z0-9._]+\.[a-zA-Z0-9._]+$/i
const pwdEx = /^[a-zA-Z0-9!@#$%^&*()_+-/.<>]{5,71}$/i

/**
 * @param {string} name
 * @throws {string} Will throw an error if the name is not valid.
 */
export function validateName(name) {
	if (name.length < 2) {
		throw new Error("Must be at least 2 characters long")
	} else if (name.length > 60) {
		throw new Error("Must not exceed 60 characters")
	} else if (!nameEx.test(name)) {
		throw new Error("Can only contain letters and numbers")
	}
}

/**
 * @param {string} email
 * @throws Will throw an error if the email is not valid.
 */
export function validateEmail(email) {
	if (email.length < 3) {
		throw new Error("Must be at least 3 characters long")
	} else if (!emailEx.test(email)) {
		throw new Error("Please, enter a valid email address")
	}
}

/**
 * @param {string} password
 * @throws Will throw an error if the password is not valid.
 */
export function validatePassword(password) {
	if (password.length < 5) {
		throw new Error("Must be at least 5 characters long")
	} else if (password.length > 71) {
		throw new Error("Must not exceed 71 characters")
	} else if (!pwdEx.test(password)) {
		throw new Error(
			"Can only contain letters, numbers, and !@#$%^&*()_+-/.<>",
		)
	}
}
