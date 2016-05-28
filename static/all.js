var london = {lat: 51.5074, lng: 0.1278};

var map;
function initMap() {
  map = new google.maps.Map(document.getElementById('map'), {
    center: london,
    zoom: 15
  });
}
