var app = new Vue({
    el: '#app',
    data: {
        streams: null
    },
    methods: {
        getStreams: function(event) {
            axios
            .get("/api/streams")
            .then(response => {
                this.streams = response.streams
            })
        },
        createStream: function(name) {
            axios
            .post("/api/streams", { "streamName": name })
            .then(response => {
                console.log(response)
                return axios.get("/api/streams")
            })
            .then(response => {
                console.log(response)
                this.streams = response.data.streams
            })
        },
    },
    beforeMount () {
        /*
        axios
        .post("/api/streams", { "streamName": "demoA" })
        .then(response => {
            console.log(response)
            return axios.post("/api/streams", { "streamName" : "demoB" })
        }).then(response => {
            console.log(response);
            return this.getStreams();
        });
        this.getStreams();
        */
        $('#myModal').on('shown.bs.modal', function () {
            $('#myInput').trigger('focus')
        })
    },
})

let pc = new RTCPeerConnection({
    iceServers: [{
        urls: [
            'stun:stun.l.google.com:19302',
        ]
    }]
})
let ws = new WebSocket("ws://" + document.location.host + "/api/ws");
ws.onmessage = function(event) {
    if (event.data.length > 0) {
        var msg = JSON.parse(event.data)
        if (msg.error) { // error case
            console.log("error: " + msg.error)
        } else {
            var answer = msg;
            console.log(answer)
            console.log("peer id : " + answer.peerId)
            pc.setRemoteDescription(new RTCSessionDescription(answer.sdp))
        }
    }
}
ws.onopen = function(){

}

pc.ontrack = event => {
    if (event.track.kind == "audio") {
        console.log("audio track added")
    }
    if (event.track.kind == "video") {
        console.log("video track added")
    }
}

pc.oniceconnectionstatechange = e => {
    console.log(pc.iceConnectionState)
    if (pc.iceConnectionState == "connected") {
        var el = document.createElement("video")
        var stream = new MediaStream()
        stream.addTrack(pc.getRemoteStreams()[0].getAudioTracks()[0])
        stream.addTrack(pc.getRemoteStreams()[0].getVideoTracks()[0])
        el.srcObject = stream
        el.autoplay = false
        el.controls = true
        el.width = 640
        el.height = 360
        document.getElementById('remoteVideos').appendChild(el)        
    }
}

// Offer to receive 1 audio, and 1 video tracks
pc.addTransceiver('audio', {'direction': 'recvonly'})
pc.addTransceiver('video', {'direction': 'recvonly'})
pc.createOffer().then(d => {
    pc.setLocalDescription(d)
})
pc.onicecandidate = event => {
    //console.log(event.candidate)
    if (event.candidate === null) {
        var data = {
            stream: "test1",
            sdp: pc.localDescription
        }
        ws.send(JSON.stringify(data))
    }
}
