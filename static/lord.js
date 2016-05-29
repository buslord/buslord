

function updateStops(bounds) {

  var params = {
    "neLat": bounds.getNorthEast().lat(), 
    "neLng": bounds.getNorthEast().lng(), 
    "swLat": bounds.getSouthWest().lat(),
    "swLng": bounds.getSouthWest().lng()
  };

  $.getJSON('/stops?' + $.param(params), function(data) {
    // TODO remove out of bound stops
    drawStops(data);
  }); 
}

function fetchETAs(stopID, callback) {
  console.log("gonna fetch etas");
  $.getJSON('/etas?stop=' + stopID, function(data) {
      callback(data);
  });
}

