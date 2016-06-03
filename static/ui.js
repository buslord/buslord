var london = {lat: 51.500246, lng: -0.127504};

var map;
var marker;

function initMap() {
  map = new google.maps.Map(document.getElementById('map'), {
    center: london,
    zoom: 16
  });

  // prevent too many requests from happening 
  map.addListener("bounds_changed", _.throttle(function() {
    updateStops(map.getBounds());
  }, 1000));

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
    map.setCenter(latLng);
  });

  watchCountdowns();
}

var currentStops = {};

function drawStops(stops) {
  var onScreenStopIDs = [];
  for (var i = 0; i < stops.length; i++) {
    var stop = stops[i];

    onScreenStopIDs.push(stop.id);

    if (currentStops[stop.id]) {
      // we already have this stop on the map
      // nothing to do 
    } else {
    // create a new marker
    var marker = new google.maps.Marker({
        position: stop.lat_lng,
        map: map,
        animation: google.maps.Animation.DROP,
        icon: "/static/bus-marker-icon.png",
        title: stop.name 
      });
      addStopClickListener(marker, stop.id, stop.name);
      currentStops[stop.id] = marker;
    }
  }

  // remove from the map the stops we didn't get
  var newStops = {};
  _.forEach(currentStops, function(marker, stopID) {
      if (_.includes(onScreenStopIDs, stopID) == false) {
        marker.setMap(null);
        delete(currentStops[stopID]);
      }
  });

}


function addStopClickListener(marker, stopID, stopName) {
  marker.addListener("click", function(e) {
    buildAndShowEtas(stopID, stopName);
  });
} 

function buildAndShowEtas(stopID, stopName) {
    $("#etas-popup").show();
    $("#etas-popup .stopName").html(stopName);
    fetchETAs(stopID, function(etas) {
      $("#etas-popup table").html("");
      $("#etas-popup table").append("<tr><th>line</th><th>arrival</th></tr>");
      for (var j = 0; j < etas.length; j++) {
          var eta = etas[j];
          if (eta.eta > 5) {
            var klass = '';
            if (j % 2 == 0) {
              klass = 'class="shade"';
            }
            $("#etas-popup table").append("<tr " + klass + "><td>" + eta.mode_name + " <b>" + eta.line_name + '</b></td><td class="countdown" stop_name="' + stopName + '" stop_id="' + stopID + '" value="'+ eta.eta +'"></td></li>');
          }
      }
    });
}

function countdownTick() {
    $(".countdown").each(function(i, el) {
      var val = $(el).attr("value");
      val -= 1;
      $(el).attr("value", val);

      var mins = Math.floor(val/60);
      var sVal = "";
      if (mins > 0) {
        sVal = mins + "m ";
      }
      var seconds = val%60;
      sVal += seconds + "s";
 
      if ((mins <= 0) && (seconds <= 0)) {
        buildAndShowEtas($(el).attr("stop_id"), $(el).attr("stop_name"));
      } else {
        $(el).html(sVal);
      }
    });
}

function watchCountdowns() {
  setTimeout(function() {
    countdownTick();
    watchCountdowns();
  }, 1000);
}

var geoWatchID; 
if ("geolocation" in navigator) {
  geoWatchID = navigator.geolocation.watchPosition(function(position) {

    var latLng = {lat: position.coords.latitude, lng: position.coords.longitude}; 

    marker.setPosition(latLng);
    map.setCenter(latLng);

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
    $("#etas-popup table").html("");
  });
});
