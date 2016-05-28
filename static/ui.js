var london = {lat: 51.5074, lng: 0.1278};

var map;
var marker;

function initMap() {
  map = new google.maps.Map(document.getElementById('map'), {
    center: london,
    zoom: 15
  });

  marker = new google.maps.Marker({
    position: london,
    map: map,
    draggable: true,
    animation: google.maps.Animation.DROP,
    title: 'You!'
  });

  marker.addListener("dragend", function() {
    var latLng = marker.getPosition();

    updateStops(latLng); 

    map.setCenter(latLng);
  });
}

function drawStops(stops) {
  // TODO this may happen before the map is loaded 
  for (var i = 0; i < stops.length; i++) {
    var stop = new google.maps.Marker({
      position: stops.latLng,
      map: map,
      icon: "/static/bus-marker-icon.png" 
    });

    stop.addListener("click", function() {
      $("#etas-popup").show();
      $("#etas-popup ul").html("");
      // TODO spinner
      fetchETAs(stop, function(etas) {
        for (var j = 0; j < etas.length; j++) {
            var eta = etas[j];
            $("#etas-popup ul").append("<li>" + eta.bus + ": " + eta.eta);
        }
      });
    });
  }
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


