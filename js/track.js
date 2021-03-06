var start = new Date().getTime();
var original_onload = null;
var original_onerror = null;
var original_onbeforeunload = null;
var tracker = null;
var host = null;

if (window.onerror != null) original_onerror = window.onerror;
if (window.onload != null) original_onload = window.onload;
if (window.onbeforeunload != null) original_onbeforeunload = window.onbeforeunload



send_data = function(type, data) {
    var e = encodeURIComponent;
    data["host"] = e(window.location.href)
    pars = "";
    for (k in data){
        pars = pars + "&"+k+"="+e(data[k])
    }
    (new Image()).src = host + tracker + '/?t=' + e(type)+ pars;
}

collect_onbeforeunload_time = function(){
            now = new Date().getTime();
            d = new Array();
            d["beforeunloadtime"] = now;
            d["start"] = start;
            d["total_elapsed_time"] = now - start;
            try {
                if (performance != null){
                    n = performance.navigation;
                    switch(n.type){
                        case n.TYPE_RELOAD:
                            d["type_navigate"] = "reload";
                            break;
                        case n.TYPE_BACK_FORWARD:
                            d["type_navigate"] = "back_forward";
                            break;
                        default:
                        case n.TYPE_NAVIGATE:
                            d["type_navigate"] = "navigate";
                            break;
                    }
                    send_data("u", d)
                }
            } catch(e){
            }
        if (original_onbeforeunload != null) original_onbeforeunload();
}

collect_performance_data = function(){
    var d;
    try {
        if (performance != null) {
            var t = performance.timing;
            var n = performance.navigation;
            if (t.loadEventEnd > 0) {
              var page_load_time = t.loadEventEnd - t.navigationStart;
              var tcp_time = t.connectEnd - t.connectStart;
              var dns_time = t.domainLookupEnd - t.domainLookupStart; 
              var processing_time = t.loadEventEnd - t.domLoading
              var d = new Array();
              if (n.type == n.TYPE_NAVIGATE || n.type == n.TYPE_RELOAD) {
                d["page_load_time"] = page_load_time;
                d["tcp_time"] = tcp_time;
                d["dns_time"] = dns_time;
                d["processing_time"] = processing_time;
              }
            }
        } else {
            now = new Date().getTime();
            var processing_time = now - start;
            d["processing_time"] = processing_time;
        }
        send_data("p", d);
    } catch(e){
    }
    if (original_onload != null) original_onload();
}

collect_errors = function(msg, url, line) {
    d = {"msg":msg, "url": url, "line": line};
    send_data("e", d);
    if (original_onerror != null) original_onerror(msg, url, line);
}

lazy_collect = function(){ setTimeout(function(){
        collect_performance_data();
    }, 0);
}

gather_geo_info = function(){
    try {
        if (navigator.geolocation){
            navigator.geolocation.getCurrentPosition(
                    function(position){ 
                        send_data("g", {"lat": position.coords.latitude, "long": position.coords.longitude}); 
                    }, function(error){});i
        }
    } catch(e){
    }
}

activate = function(h, id) {
    host = h;
    tracker = id;
    geo = false;
    window.onload = lazy_collect
    window.onerror = collect_errors;
    window.onbeforeunload = collect_onbeforeunload_time;
    if (geo == true) { gather_geo_info(); }
}
