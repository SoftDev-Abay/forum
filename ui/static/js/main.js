var navLinks = document.querySelectorAll("nav a");
for (var i = 0; i < navLinks.length; i++) {
	var link = navLinks[i]
	if (link.getAttribute('href') == window.location.pathname) {
		link.classList.add("live");
		break;
	}
}

  const openBtn = document.getElementById("openReportModal");
  const modal = document.getElementById("reportWindow");
  const closeBtn = document.getElementById("closeModal");

  openBtn?.addEventListener("click", function() {
    modal.style.display = "block";
  });

  closeBtn?.addEventListener("click", function() {
    modal.style.display = "none";
  });

  window.addEventListener("click", function(e) {
    if (e.target === modal) {
      modal.style.display = "none";
    }
  });