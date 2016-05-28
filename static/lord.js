

function updateStops(latLng) {
  fetch('/stops?' + $.param(latLng)).then(function(response) {
    // TODO remove out of bound stops

    drawStops(response.json());
  }); 
}

function fetchETAs(stop, callback) {
  fetch('/etas?stop=' + stop.id).then(function(response) {
    if (callback) {
      callback(response.json());
    }
  });
}


$(document).ready(function() {
  console.log("loaded");

  updateStops(london);

});
