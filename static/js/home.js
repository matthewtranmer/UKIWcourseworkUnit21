window.addEventListener('resize', splash);
splash()

function splash(e){
    var width = (window.innerWidth > 0) ? window.innerWidth : screen.width;
    document.getElementById("splashtext").setAttribute("style", "font-size: " + width/2.7 +"%;")

    if (width > 1000){
        document.getElementById("splashimg").setAttribute("src", "/static/img/homesplash.jpg")
    }
    else{
        document.getElementById("splashimg").setAttribute("src", "/static/img/smallersplash.jpg")
    }
    
}
