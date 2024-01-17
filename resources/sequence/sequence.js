

let isPressed = false
let startTime = new Date();
const silencePause = 2000;

let currentHits = [];
let sequences = [];


class Hit {
    value
    silence
    constructor(value,silence) {
        this.value = value;
        this.silence = silence;
    }
    total(){
        return this.value + this.silence;
    }
}

let previousSilence = 0;
const down = e => {
    if(isPressed){
        return
    }
    previousSilence = new Date() - startTime;
    if(previousSilence > silencePause && currentHits.length > 0){
        sequences.unshift(currentHits)
        currentHits = [];
    }
    isPressed = true;
    startTime = new Date();
}

const up = e => {
    let silence = currentHits.length > 0 ? previousSilence : 0;
    currentHits.push(new Hit(new Date() - startTime, silence))
    isPressed = false;
    startTime = new Date();
}

const analyze = () => {
   const notes = extractNotes()
    switch(notes.length){
        case 0 : console.log("erreur");break;
        case 1 : console.log("TURN ON");break;
        case 2 : console.log("NOTHING");break;
        default : detectPattern(notes)
    }
}

const patterns = {}

const detectPattern = notes => {
    const line = notes.join('');
    console.log(line)
}

const extractNotes = () => {
    const seq = sequences.pop();
    if (seq == null){
        return [];
    }
    // Compare all silence to the first
    // First is a black, compute everything
    const reference = seq[1].total()
    const notes = []
    for(let i = 2 ; i < seq.length ; i++){
        const ratio = reference / seq[i].total();
        notes.push(test(ratio))
    }
    return notes;
}

const test = (value) => {
    for (let i = 1 ; i <=8 ; i++) {
        const v = value / i;
        if (v > 0.7 && v < 1.3){
            return i;
        }
    }
    return -1
}

function init() {
    document.addEventListener('keydown',down)
    document.addEventListener('keyup',up)
    document.getElementById('show_sequence').onclick = () => console.log(sequences)
    document.getElementById('analyze_sequence').onclick = () => analyze()
    setInterval(()=>{
        // check last previous silence, push in sequence if necessary
        if((new Date() - startTime) > silencePause && currentHits.length > 0){
            sequences.unshift(currentHits)
            currentHits = [];
        }
    },200)
}

init();