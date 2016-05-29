var london = {lat: 51.5074, lng: 0.1278};

var map;
var marker;

function initMap() {
  map = new google.maps.Map(document.getElementById('map'), {
    center: london,
    zoom: 15
  });

  updateStops(london);

  marker = new google.maps.Marker({
    position: london,
    map: map,
    draggable: true,
    animation: google.maps.Animation.DROP,
    title: 'You!'
  });

  marker.addListener("dragend", function() {
    var pos = marker.getPosition();
    var latLng = {lat: pos.lat(), lng: pos.lng()};
    updateStops(latLng); 
    map.setCenter(latLng);
  });
}

function drawStops(stops) {
  for (var i = 0; i < stops.length; i++) {
    var stop = stops[i];
    var marker = new google.maps.Marker({
      position: stop.lat_lng,
      map: map,
      animation: google.maps.Animation.DROP,
      icon: "/static/bus-marker-icon.png" 
    });

    addStopClickListener(marker, stop.id);
  }
}

function addStopClickListener(marker, stopID) {
  marker.addListener("click", function(e) {
    $("#etas-popup").show();
    $("#etas-popup ul").html("");
    fetchETAs(stopID, function(etas) {
      for (var j = 0; j < etas.length; j++) {
          var eta = etas[j];
          $("#etas-popup ul").append("<li>" + eta.bus_name + ": " + eta.eta);
      }
    });
  });
} 


var geoWatchID; 
if ("geolocation" in navigator) {
  geoWatchID = navigator.geolocation.watchPosition(function(position) {

    var latLng = {lat: position.coords.latitude, lng: position.coords.longitude}; 

    updateStops(latLng);

    map.setCenter(latLng);
    marker.setPosition(latLng);

    $("#welcome-popup").hide();

  }, function(e) {
    console.log("some err: " + e.message);
    $("#welcome-popup").hide();
  });
} else {
  alert("no geolocation feats on this browser");
}

$(function() {
  $('#close-etas').on("click", function(e) {
    e.preventDefault();
    $("#etas-popup").hide();
  });
});
