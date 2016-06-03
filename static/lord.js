

function updateStops(bounds) {

  var params = {
    "neLat": bounds.getNorthEast().lat(), 
    "neLng": bounds.getNorthEast().lng(), 
    "swLat": bounds.getSouthWest().lat(),
    "swLng": bounds.getSouthWest().lng()
  };

  // prevent from doing too many calls. cancel the previous and take the last
  $.getJSON('/stops?' + $.param(params), function(data) {
    // remove out of bound stops
    drawStops(data);
  }); 
}

function fetchETAs(stopID, callback) {
  $.getJSON('/etas?stop=' + stopID, function(data) {
      callback(data);
  });
}

