$( document ).ready(function() {
  if ($("#slideshow")) {
    startSlideshow();
  }
});

window.interval = 5000;
window.lastRequest = 0;
window.slideType = 0;

function startSlideshow() {
  window.getImage();
  window.setInterval("window.intervalCheck()", 1 * 100);
}

function intervalCheck() {
  var now = Date.now();
  if (now - window.lastRequest > window.interval) {
    getImage();
  }

}

function getImage() {
  window.lastRequest = Date.now();

  var type = document.location.search.split("type=")[1]

  $.ajax({
  url: "/api",
  data: {"type": type},
  success: function(res) {
    renderImage(res.path)
  }
  ,
  dataType: "json"
});
}

function renderImage(path) {
  $("#slideshow img").attr("src", path);
}