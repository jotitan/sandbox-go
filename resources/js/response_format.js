function checkProgression(value){
    return value != null && value != 0 && value != 100
}

function showIfTrue(value){
    return value == true
}

function noShow(value){
    return ""
}

function modifyProgressBar(element,value,beginTime){
    element.progressbar({value:value})

    if(beginTime != null){
        var value = Math.round(new Date().getTime()/1000 - beginTime);
        element.attr('title','Launch ' + formatDuration(value) + ' ago')
    }
}

function formatDuration(value){
    var str = "";
    if(value > 3600){
        var nbHour = Math.floor(value/3600);
        str+= nbHour + "h ";
        value -= nbHour*3600;
    }
    if(value > 60){
        var nbMin = Math.floor(value/60);
        str+= nbMin + "m ";
        value -= nbMin*60;
    }
    if(value > 0){
        str+=value + "s";
    }
    return str;
}

var StatusType = ["Down","Up","No ID","Synchronizing","Waiting topology"]
function getStatus(status){
    return StatusType[status]
}

var memNames = ["Ko","Mo","Go","To","Po","Fo"]
function formatMemory(mem){
    loop = 0
    while(mem > 1024){
        mem=parseInt(mem*10/1024)/10
        loop++
    }
    return mem + " " + memNames[loop]
}

function formatCheck(check){
    switch(check){
    case 1 : return "KO"
    case 2 : return "OK"
    }
    return "-"
}

var ClassType = ["node_down","node_up","node_no_id","node_synchro","node_wait_topo"]

function getClassStatus(status){
    return ClassType[status];
}