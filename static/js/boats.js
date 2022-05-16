var boat_div = document.getElementById("boats");

window.addEventListener('resize', resize);
resize();
function resize(e){
    var screen_width = (window.innerWidth > 0) ? window.innerWidth : screen.width;

    boat_div.style.display = "inline";
    var div_width = boat_div.style.width;
    boat_div.style.display = "block";

    boat_div.style.marginLeft = (screen_width - div_width)/2;
    console.log(div_width);
}
