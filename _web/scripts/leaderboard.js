;(() => {
	if (window.location.pathname != "/leaderboard") return

	for (const div of document.getElementsByClassName("leaderboard-row-time")) {
		div.textContent = new Date(div.textContent).toLocaleDateString(
			"en-US",
			{
				month: "short",
				day: "2-digit",
				year: "numeric",
			},
		)
	}
})()
