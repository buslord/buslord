

function updateStops(latLng) {
  $.getJSON('/stops?' + $.param(latLng), function(data) {
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

