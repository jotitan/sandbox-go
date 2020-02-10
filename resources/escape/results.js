//HiddenWord

function searchAnagram(data,test){
Tools.clear();
let a = [];
for(let i = 0 ; i < data.length ; i++){
    for(let j = 0 ; j < data[i].length ; j++){
        a.push(data[i][j])
    }
}

let words = look(a.filter(l=>l.color==='red').map(l=>l.letter),"",[])
console.log("Found",Object.keys(words).length)
Object.keys(words).forEach(w=>test(w).then(a=>a?console.log("Trouve",w):null))
}

function look(letters,current,done) {
    let words = {};
    for (let i = 0; i < letters.length; i++) {
        if (done[i] == null) {
            let v = current + letters[i];
            if (v.length == letters.length) {
                words[v] = true;
            } else {
                let nDone = [...done];
                nDone[i] = i;
                let foundWords = look(letters, v, nDone);
                Object.keys(foundWords).forEach(w => words[w] = true);

            }
        }
    }
    return words;
}

function showLetters(data) {
    Tools.cleanCanvas();
    Tools.clear();
    for (let i = 0; i < data.length; i++) {
        for (let j = 0; j < data[i].length; j++) {
            Tools.text(10 + i * 25, 20 + j * 25, data[i][j].letter, data[i][j].color == "black" ? "white" : "red", 12);
        }
    }
}

function escargot(){
    Tools.cleanCanvas();
    Tools.clear();

    let rayon = 1;
    let angle = 0;
    let center = {x:200,y:200};
    let step = Math.PI/28;
    for (let i = 0 ; i < 8000 ; i++){
        if(i%30 === 0){
            rayon++;
            step/=1.025;
        }
        angle = (angle+(step))%(2*Math.PI);
        let x = rayon * Math.cos(angle);
        let y = rayon * Math.sin(angle);
        Tools.plot(center.x + x,center.y + y,1,"black");
    }
    test()
}
