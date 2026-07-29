// header.js makes site header responsive. On smaller screens, menu items are
// hidden in the sidebar and are displayed on a click.
const sidebar = document.getElementById("siteSidebar")
const toggle = document.getElementById("toggleSidebar")
toggle.onclick = () => {
  sidebar.classList.toggle("active")

  toggle.style.backgroundImage = sidebar.classList.contains("active")
    ? `url("/assets/images/hide-sidebar.svg")`
    : `url("/assets/images/show-sidebar.svg")`
}
