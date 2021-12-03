import { sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';
import tcp from 'k6/x/tcp';
import { randomInterval } from '../tools/random.js';
import { msgArray } from '../unit/msg_unit.js';

const strAddr = '127.0.0.1:9000';
// const strAddr = '10.89.130.4:9000';

// 总体统计数据
let gRate = new Rate('GlobalRate');
let gCounter = new Counter('GlobalCount')
let gTrend = new Trend('GlobalDuration')
let gTimeOutCounter = new Counter('GlobalTimeOutCount')

// 消息独立统计数据
let trendMap = new Map()
let counterMap = new Map()
let rateMap = new Map()
msgArray.forEach(function (e) {
    trendMap.set(e.name, new Trend(e.name + "_duration"))
    counterMap.set(e.name, new Counter(e.name + "_count"))
    rateMap.set(e.name, new Rate(e.name + "_rate"))
});

// let userName = randStr();
let client = new tcp.Client();
let bCreate = false;

export let options = {
    stages: [
        { duration: '5s', target: 800 },
        { duration: '60s', target: 800 },
        { duration: '5s', target: 0 },
    ],
}

export function setup() {
    // console.log('init test unit array');

    const msg_weight = [];
    for (const msg of msgArray) {
        for (let i = 0; i < msg.weight; i++) {
            msg_weight.push(msg.name);
        }
    }
    return msg_weight;
}

export default function (msg_weight) {
    getData()
    sendMsg(msg_weight)
    sleep(randomInterval(1, 1));
}

function sendMsg(msg_weight) {
    if (!bCreate) {
        let err = client.connect(strAddr, OnCallBackFromGolang)

        if (err == null) {
            bCreate = true;
            client.sendMsgToServer("C2GVerifyMessage", true)
        } else {
            console.info("Failed to create conn.")
        }
    }
    else if (Boolean(client.connValid())) {
        let randIdx = randomInterval(0, msg_weight.length);
        let msgName = msg_weight[randIdx];
        client.sendMsgToServer(msgName)
    }
}

function getData() {
    let datas = client.getMsgData();
    let timeout = false
    for (let data of datas) {
        // console.info(String(data.cmd), Boolean(data.sus), Number(data.duration))

        if (trendMap.has(String(data.cmd))) {
            trendMap.get(String(data.cmd)).add(Number(data.duration))
            if (Number(data.duration) > 2900) { timeout = true }
            gTrend.add(Number(data.duration))
        }

        if (counterMap.has(String(data.cmd))) {
            counterMap.get(String(data.cmd)).add(1)
            gCounter.add(1)
        }

        if (rateMap.has(String(data.cmd))) {
            rateMap.get(String(data.cmd)).add(Boolean(data.sus))
            gRate.add(Boolean(data.sus))
        }
    }

    if (timeout) {
        console.log(client.getID() + ", msg timeout.")
        gTimeOutCounter.add(1)
        client.closeConn()
    }
}

export function OnCallBackFromGolang(msg, sus) {
    // console.info(String(client.getID()), String(msg), Boolean(sus))
    // globalRate.add(Boolean(sus))

    // let m = String(msg)
    // if (trendMap.has(m)) {
    //     trendMap.get(m).add(1)
    // }
    // if (counterMap.has(m)) {
    //     counterMap.get(m).add(1)
    // }
}

export function teardown() {
    // Cleanly shutdown and flush telemetry when k6 exits.
}
