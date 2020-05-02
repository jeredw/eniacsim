function getTransformToElement(elem, toElement) {
  return toElement.getScreenCTM().inverse().multiply(elem.getScreenCTM());
};

function transformedBoundingBox(elem) {
  // https://stackoverflow.com/questions/10623809/get-bounding-box-of-element-accounting-for-its-transform
  const svg = elem.ownerSVGElement;
  const t = getTransformToElement(elem, svg);

  let ps = [svg.createSVGPoint(), svg.createSVGPoint(), svg.createSVGPoint(), svg.createSVGPoint()];
  let box = elem.getBBox();
  [ps[0].x, ps[0].y] = [box.x, box.y];
  [ps[1].x, ps[1].y] = [box.x + box.width, box.y];
  [ps[2].x, ps[2].y] = [box.x + box.width, box.y + box.height];
  [ps[3].x, ps[3].y] = [box.x, box.y + box.height];

  let xMin = Infinity;
  let xMax = -Infinity;
  let yMin = Infinity;
  let yMax = -Infinity;
  for (const p of ps) {
    const pt = p.matrixTransform(t);
    xMin = Math.min(xMin, pt.x);
    xMax = Math.max(xMax, pt.x);
    yMin = Math.min(yMin, pt.y);
    yMax = Math.max(yMax, pt.y);
  }

  box.x = xMin;
  box.y = yMin;
  box.width = xMax - xMin;
  box.height = yMax - yMin;
  return box;
}

let source = new EventSource('/events');
let eniac = null;
let defaultViewBox = '';
let neons = [];
let machineState = {};

source.addEventListener('message', (event) => {
  machineState = JSON.parse(event.data);
  machineState.pulse = parseInt(machineState.cycling, 10) / 2;
});

function step(ts) {
  for (const neon of neons) {
    neon(machineState);
  }
  requestAnimationFrame(step);
}

function connectNeon(selector, isTurnedOn) {
  const neon = eniac.querySelector(selector);
  const onColor = '#ffd43a';
  const offColor = '#574400';
  neon.style.contain = 'layout paint';
  neon.style.fill = offColor;
  neons.push((s) => {
    neon.style.fill = isTurnedOn(s) ? onColor : offColor;
  });
}

function addStylesToSvg(doc) {
  let style = document.createElement('style');
  doc.documentElement.appendChild(style);
  style.sheet.insertRule('svg text { user-select: none; }');
}

let adjustRotary = null;

function connectRotarySwitch(selector, settings, onChange=undefined) {
  const rotary = eniac.querySelector(selector);
  const svg = rotary.ownerSVGElement;
  rotary.style.contain = 'layout paint';
  let rotation = svg.createSVGTransform();
  let [cx, cy] = [0, 0];
  if (!rotary.classList.contains('knub')) {
    const wiper = rotary.querySelector('path');
    const box = wiper.getBBox();
    [cx, cy] = [box.x + box.width / 2, box.y + box.height / 2];
    rotation.setRotate(0, cx, cy);
    wiper.transform.baseVal.appendItem(rotation);
    wiper.style.transition = 'transform 100ms linear';
  } else {
    const base = rotary.querySelector('circle');
    const box = base.getBBox();
    [cx, cy] = [box.x + box.width / 2, box.y + box.height / 2];
    rotation.setRotate(0, cx, cy);
    rotary.transform.baseVal.appendItem(rotation);
    rotary.style.transition = 'transform 100ms linear';
  }
  let index = settings.findIndex(s => s.degrees == 0);
  rotary.dataset.value = settings[index].value;
  rotary.addEventListener('click', (event) => {
    event.stopPropagation();
    const delta = event.metaKey ? -1 : 1;
    index += delta;
    if (index >= settings.length) {
      index = 0;
    }
    if (index < 0) {
      index = settings.length - 1;
    }
    const newValue = settings[index].value;
    rotary.dataset.value = newValue;
    if (onChange) {
      onChange(newValue);
    }
    rotation.setRotate(settings[index].degrees, cx, cy);
    // for manual adjustments
    adjustRotary = (degrees) => {
      rotation.setRotate(degrees, cx, cy);
    };
  });
}

function connectToggleSwitch(selector) {
  const toggle = eniac.querySelector(selector);
  const svg = toggle.ownerSVGElement;
  const switchBody = toggle.querySelector('g');
  const base = toggle.querySelector('circle');
  const box = base.getBBox();
  const [cx, cy] = [box.x + box.width / 2, box.y + box.height / 2];
  toggle.style.contain = 'layout paint';
  switchBody.style.transition = 'transform 100ms linear';
  let translate = svg.createSVGTransform();
  let untranslate = svg.createSVGTransform();
  let scale = svg.createSVGTransform();
  translate.setTranslate(cx, cy);
  scale.setScale(1.0, 1.0);
  untranslate.setTranslate(-cx, -cy);
  switchBody.transform.baseVal.appendItem(translate);
  switchBody.transform.baseVal.appendItem(scale);
  switchBody.transform.baseVal.appendItem(untranslate);
  toggle.dataset.value = 'on';
  toggle.addEventListener('click', (event) => {
    event.stopPropagation();
    if (toggle.dataset.value == 'on') {
      toggle.dataset.value = 'off';
      scale.setScale(-1.0, 1.0);
    } else {
      toggle.dataset.value = 'on';
      scale.setScale(1.0, 1.0);
    }
  });
}

function connectButton(selector) {
  const button = eniac.querySelector(selector);
  const svg = button.ownerSVGElement;
  const box = button.getBBox();
  const [cx, cy] = [box.x + box.width / 2, box.y + box.height / 2];
  let translate = svg.createSVGTransform();
  let untranslate = svg.createSVGTransform();
  let scale = svg.createSVGTransform();
  translate.setTranslate(cx, cy);
  scale.setScale(1.0, 1.0);
  untranslate.setTranslate(-cx, -cy);
  button.transform.baseVal.appendItem(translate);
  button.transform.baseVal.appendItem(scale);
  button.transform.baseVal.appendItem(untranslate);
  button.dataset.value = 'off';
  button.addEventListener('mousedown', (event) => {
    event.stopPropagation();
    button.dataset.value = 'on';
    scale.setScale(0.8, 0.8);
  });
  button.addEventListener('mouseup', (event) => {
    event.stopPropagation();
    button.dataset.value = 'off';
    scale.setScale(1.0, 1.0);
  });
}

function makeNeedleRotateable(selector) {
  const needle = eniac.querySelector(selector);
  const svg = needle.ownerSVGElement;
  const box = needle.getBBox();
  const [rx, ry] = [box.x + box.width, box.y + box.height];
  needle.style.transition = 'transform 100ms linear';
  let rotation = svg.createSVGTransform();
  rotation.setRotate(0, rx, ry);
  needle.transform.baseVal.appendItem(rotation);
  return (angle) => rotation.setRotate(angle, rx, ry);
}

function makePanelSelectable(panelSelector) {
  const panel = eniac.querySelector(panelSelector);
  // TODO: hover feedback and right click tip
  panel.addEventListener('click', (event) => {
    viewElementSelector(panelSelector);
  });
  panel.addEventListener('contextmenu', (event) => {
    event.preventDefault();
    viewDefault();
  });
}

function connectInitiateElements() {
  makePanelSelectable('#initiate-panel');
  const rotateNeedle1 = makeNeedleRotateable('#initiate-vm1-needle');
  const showRandomTrace = () => {
    const traces = eniac.querySelectorAll('.initiate-trace');
    const choice = ~~(Math.random() * traces.length);
    for (let i = 0; i < traces.length; i++) {
      traces[i].style.visibility = i == choice ? '' : 'hidden';
    }
    rotateNeedle1(Math.random() * 73);
  };
  connectRotarySwitch('#initiate-dcv1', [
    {value: 'A', degrees: -150},
    {value: 'B', degrees: -110},
    {value: 'C', degrees: -80},
    {value: 'D', degrees: -58},
    {value: 'E', degrees: -25},
    {value: 'F', degrees: 0},
    {value: 'G', degrees: 20},
    {value: 'H', degrees: 54},
  ], showRandomTrace);
  connectRotarySwitch('#initiate-dcv2', [
    {value: '1', degrees: -185},
    {value: '2', degrees: -152},
    {value: '3', degrees: -116},
    {value: '4', degrees: -82},
    {value: '5', degrees: -54},
    {value: '6', degrees: -29},
    {value: '7', degrees: 0},
    {value: '8', degrees: 24},
    {value: '9', degrees: 53},
    {value: '10', degrees: 84},
    {value: '11', degrees: 117},
  ], showRandomTrace);
  const rotateNeedle2 = makeNeedleRotateable('#initiate-vm2-needle');
  connectRotarySwitch('#initiate-acv', [
    {value: 'VAB', degrees: -115},
    {value: 'VBC', degrees: -61},
    {value: 'VCA', degrees: 0},
  ], () => rotateNeedle2(Math.random() * 73));
  connectButton('#initiate-start');
  connectButton('#initiate-stop');
  connectButton('#initiate-rs');
  connectButton('#initiate-clear');
  connectButton('#initiate-door');
  connectButton('#initiate-pulse');

  makePanelSelectable('#initiate-bottom');
  connectNeon('#initiate-neon-sc1',   (s) => s.initiate[0] == '1');
  connectNeon('#initiate-neon-sc2',   (s) => s.initiate[1] == '1');
  connectNeon('#initiate-neon-sc3',   (s) => s.initiate[2] == '1');
  connectNeon('#initiate-neon-sc4',   (s) => s.initiate[3] == '1');
  connectNeon('#initiate-neon-sc5',   (s) => s.initiate[4] == '1');
  connectNeon('#initiate-neon-sc6',   (s) => s.initiate[5] == '1');
  connectNeon('#initiate-neon-rs',    (s) => s.initiate[6] == '1');
  connectNeon('#initiate-neon-ps',    (s) => s.initiate[7] == '1');
  connectNeon('#initiate-neon-rf',    (s) => s.initiate[8] == '1');
  connectNeon('#initiate-neon-ri',    (s) => s.initiate[9] == '1');
  connectNeon('#initiate-neon-rsync', (s) => s.initiate[10] == '1');
  connectNeon('#initiate-neon-pf',    (s) => s.initiate[11] == '1');
  connectNeon('#initiate-neon-psync', (s) => s.initiate[12] == '1');
  connectNeon('#initiate-neon-ip',    (s) => s.initiate[13] == '1');
  connectNeon('#initiate-neon-isync', (s) => s.initiate[14] == '1');
}

function connectCyclingElements() {
  makePanelSelectable('#cycling-top');
  for (let i = 1; i <= 20; i++) {
    connectNeon(`#cycling-neon-r${i}`, ((i) => {
      return (s) => s.pulse == i-1;
    })(i));
  }
  connectNeon('#cycling-neon-10p', () => false);
  connectNeon('#cycling-neon-ccg', () => false);

  makePanelSelectable('#cycling-panel');
  connectToggleSwitch('#cycling-heater-toggle');
  const showPulseTrace = (value) => {
    const traces = eniac.querySelectorAll('.cycling-trace');
    const selected = eniac.querySelector('#cycling-trace-' + value);
    for (const trace of traces) {
      trace.style.visibility = trace == selected ? '' : 'hidden';
    }
  };
  connectRotarySwitch('#cycling-osc', [
    {value: 'ext', degrees: -205},
    {value: 'cpp', degrees: -180},
    {value: '10p', degrees: -150},
    {value: '9p', degrees: -115},
    {value: '1p', degrees: -80},
    {value: '2p', degrees: -60},
    {value: "2pp", degrees: -30},
    {value: "4p", degrees: 0},
    {value: "1pp", degrees: 30},
    {value: "ccg", degrees: 55},
    {value: "9p", degrees: 75},
    {value: "scg", degrees: 100},
  ], showPulseTrace);
  connectRotarySwitch('#cycling-op', [
    {value: '1p', degrees: -60},
    {value: '1a', degrees: -30},
    {value: 'co', degrees: 0},
  ]);
  connectButton('#cycling-p');

  makePanelSelectable('#cycling-bottom');
  connectRotarySwitch('#cycling-osc-source', [
    {value: 'I', degrees: 0},
    {value: 'E', degrees: 100},
  ]);
}

function connectMPElements(panelNumber) {
  const prefix = `mp${panelNumber}`;
  const steppers = panelNumber == 1 ? 'abcde' : 'fghjk';
  const startDecade = panelNumber == 1 ? 20 : 10;
  const baseStepper = panelNumber == 1 ? 0 : 5;

  makePanelSelectable(`#${prefix}-top`);
  for (let decade = startDecade; decade > startDecade - 10; decade--) {
    for (let value = 9; value >= 0; value--) {
      connectNeon(`#${prefix}-neon-d${decade}v${value}`, ((d, v) => {
        return (s) => s.mp.decade[20-d] == v;
      })(decade, value));
    }
  }
  for (let i = 0; i < steppers.length; i++) {
    for (let value = 6; value >= 1; value--) {
      connectNeon(`#${prefix}-neon-s${steppers[i]}${value}`, ((i, v) => {
        return (s) => s.mp.stage[i] == v;
      })(baseStepper + i, value));
    }
  }

  makePanelSelectable(`#${prefix}-panel`);
  connectToggleSwitch(`#${prefix}-heater-toggle`);
  for (let i = 0; i < steppers.length - 1; i++) {
    const [s1, s2] = [steppers[i], steppers[i+1]];
    connectRotarySwitch(`#${prefix}-assoc-${s1}${s2}`, [
      {value: s1, degrees: -35},
      {value: s2, degrees: 0},
    ]);
  }
  const decadeSettings = [
    {value: '0', degrees: -185},
    {value: '1', degrees: -155},
    {value: '2', degrees: -115},
    {value: '3', degrees: -85},
    {value: '4', degrees: -55},
    {value: '5', degrees: -25},
    {value: '6', degrees: 0},
    {value: '7', degrees: 25},
    {value: '8', degrees: 55},
    {value: '9', degrees: 85},
  ];
  for (let decade = startDecade; decade > startDecade - 10; decade--) {
    for (let digit = 1; digit <= 6; digit++) {
      connectRotarySwitch(`#${prefix}-d${decade}s${digit}`, decadeSettings);
    }
  }
  const clearSettings = [
    {value: '1', degrees: -110},
    {value: '2', degrees: -80},
    {value: '3', degrees: -55},
    {value: '4', degrees: -30},
    {value: '5', degrees: 0},
    {value: '6', degrees: 20},
  ];
  for (const s of steppers) {
    connectRotarySwitch(`#${prefix}-c${s}`, clearSettings);
  }

  makePanelSelectable(`#${prefix}-bottom`);
  for (let i = 0; i < steppers.length; i++) {
    connectNeon(`#${prefix}-neon-${steppers[i]}in`, ((i) => {
      return (s) => s.mp.inff[i] > 0;
    })(baseStepper + i));
  }
}

function connectFT1Elements(ftNumber) {
  const unit = `#ft${ftNumber}-1`;
  makePanelSelectable(`${unit} .top-panel`);
  for (let d = 0; d <= 9; d++) {
    connectNeon(`${unit} .arg-x${d}`, ((d) => {
      return (s) => (s.ft[ftNumber-1].arg%10) == d;
    })(d));
  }
  for (let d = 0; d <= 10; d++) {
    connectNeon(`${unit} .arg-${d}x`, ((d) => {
      return (s) => ~~(s.ft[ftNumber-1].arg/10) == d;
    })(d));
  }
  connectNeon(`${unit} .arg-setup`, (s) => s.ft[ftNumber-1].argSetup);
  connectNeon(`${unit} .add`, (s) => s.ft[ftNumber-1].add);
  connectNeon(`${unit} .subtract`, (s) => s.ft[ftNumber-1].subtract);
  for (let i = 0; i <= 12; i++) {
    connectNeon(`${unit} .ring${i}`, ((i) => {
      return (s) => s.ft[ftNumber-1].ring == i;
    })(i));
  }

  makePanelSelectable(`${unit} .front-panel`);
  connectToggleSwitch(`${unit} .heater-toggle`);
  const opSettings = [
    {value: 'A-2', degrees: -185},
    {value: 'A-1', degrees: -150},
    {value: 'A0', degrees: -120},
    {value: 'A+1', degrees: -90},
    {value: 'A+2', degrees: -65},
    {value: 'S+2', degrees: -30},
    {value: 'S+1', degrees: 0},
    {value: 'S0', degrees: 25},
    {value: 'S-1', degrees: 55},
    {value: 'S-2', degrees: 85},
  ];
  const clearSettings = [
    {value: '0', degrees: 0},
    {value: 'NC', degrees: 25},
    {value: 'C', degrees: 55},
  ];
  const repeatSettings = [
    {value: '1', degrees: -185},
    {value: '2', degrees: -150},
    {value: '3', degrees: -115},
    {value: '4', degrees: -90},
    {value: '5', degrees: -60},
    {value: '6', degrees: -25},
    {value: '7', degrees: 0},
    {value: '8', degrees: 30},
    {value: '9', degrees: 60},
  ];
  for (let i = 1; i <= 11; i++) {
    connectRotarySwitch(`${unit} .op${i}`, opSettings);
    connectRotarySwitch(`${unit} .cl${i}`, clearSettings);
    connectRotarySwitch(`${unit} .rp${i}`, repeatSettings);
  }

  makePanelSelectable(`${unit} .bottom-panel`);
  for (let p = 0; p <= 10; p++) {
    connectNeon(`${unit} .inff${p}`, ((p) => {
      return (s) => s.ft[ftNumber-1].inff[p];
    })(p));
  }
}

function connectFT2Elements(ftNumber) {
  const unit = `#ft${ftNumber}-2`;
  makePanelSelectable(`${unit} .front-panel`);
  connectToggleSwitch(`${unit} .heater-toggle`);
  const pmSettings = [
    {value: 'P', degrees: -60},
    {value: 'M', degrees: -30},
    {value: 'T', degrees: 0},
  ];
  connectRotarySwitch(`${unit} .mpm1`, pmSettings);
  connectRotarySwitch(`${unit} .mpm2`, pmSettings);
  const deleteSettings = [
    {value: 'D', degrees: 0},
    {value: 'O', degrees: 50},
  ];
  const constantSettings = [
    {value: '0', degrees: -215},
    {value: '1', degrees: -180},
    {value: '2', degrees: -150},
    {value: '3', degrees: -120},
    {value: '4', degrees: -95},
    {value: '5', degrees: -60},
    {value: '6', degrees: -30},
    {value: '7', degrees: 0},
    {value: '8', degrees: 20},
    {value: '9', degrees: 55},
    {value: 'PM1', degrees: 80},
    {value: 'PM2', degrees: 110},
  ];
  for (let i = 1; i <= 4; i++) {
    connectRotarySwitch(`${unit} .a${i}d`, deleteSettings);
    connectRotarySwitch(`${unit} .b${i}d`, deleteSettings);
    connectRotarySwitch(`${unit} .a${i}c`, constantSettings);
    connectRotarySwitch(`${unit} .b${i}c`, constantSettings);
  }
  const subtractSettings = [
    {value: '0', degrees: -50},
    {value: 'S', degrees: 0},
  ];
  for (let i = 5; i <= 10; i++) {
    connectRotarySwitch(`${unit} .a${i}s`, subtractSettings);
    connectRotarySwitch(`${unit} .b${i}s`, subtractSettings);
  }
}

function connectAccumulatorElements(accNumber) {
  const unit = `#accumulator-${accNumber}`;
  makePanelSelectable(`${unit} .top-panel`);
  connectNeon(`${unit} .neon-p`, (s) => !s.acc[accNumber-1].sign);
  connectNeon(`${unit} .neon-m`, (s) => s.acc[accNumber-1].sign);
  for (let decade = 10; decade >= 1; decade--) {
    connectNeon(`${unit} .d${decade}`, ((d) => {
      return (s) => s.acc[accNumber-1].decff[d-1];
    })(accNumber, decade));
    for (let value = 0; value <= 9; value++) {
      connectNeon(`${unit} .d${decade}v${value}`, ((d, v) => {
        return (s) => s.acc[accNumber-1].decade[d-1] == v;
      })(decade, value));
    }
  }

  makePanelSelectable(`${unit} .front-panel`);
  connectToggleSwitch(`${unit} .heater-toggle`);
  connectRotarySwitch(`${unit} .sf`, [
    {value: '10', degrees: -185},
    {value: '9', degrees: -150},
    {value: '8', degrees: -125},
    {value: '7', degrees: -95},
    {value: '6', degrees: -65},
    {value: '5', degrees: -35},
    {value: '4', degrees: 0},
    {value: '3', degrees: 30},
    {value: '2', degrees: 60},
    {value: '1', degrees: 90},
    {value: '0', degrees: 120},
  ]);
  connectRotarySwitch(`${unit} .sc`, [
    {value: 'C', degrees: 0},
    {value: '0', degrees: -50},
  ]);
  for (let i = 1; i <= 12; i++) {
    connectRotarySwitch(`${unit} .op${i}`, [
      {value: 'α', degrees: -150},
      {value: 'β', degrees: -125},
      {value: 'γ', degrees: -90},
      {value: 'δ', degrees: -65},
      {value: 'ε', degrees: -32},
      {value: '0', degrees: 0},
      {value: 'A', degrees: 30},
      {value: 'AS', degrees: 55},
      {value: 'S', degrees: 85},
    ]);
    connectRotarySwitch(`${unit} .cc${i}`, [
      {value: 'C', degrees: 0},
      {value: '0', degrees: -75},
    ]);
  }
  for (let i = 5; i <= 12; i++) {
    connectRotarySwitch(`${unit} .rp${i}`, [
      {value: '1', degrees: -185},
      {value: '2', degrees: -150},
      {value: '3', degrees: -115},
      {value: '4', degrees: -90},
      {value: '5', degrees: -60},
      {value: '6', degrees: -25},
      {value: '7', degrees: 0},
      {value: '8', degrees: 30},
      {value: '9', degrees: 60},
    ]);
  }

  makePanelSelectable(`${unit} .bottom-panel`);
  for (let i = 1; i <= 12; i++) {
    connectNeon(`${unit} .p${i}`, ((i) => {
      return (s) => s.acc[accNumber-1].program[i-1];
    })(i));
  }
  for (let i = 1; i <= 9; i++) {
    connectNeon(`${unit} .rep${i}`, ((v) => {
      return (s) => s.acc[accNumber-1].repeat == v;
    })(i));
  }
}

function connectDividerElements() {
  makePanelSelectable('#div-top');
  connectNeon('#div-divff',    (s) => s.div.ffs[0] == '1');
  connectNeon('#div-clrff',    (s) => s.div.ffs[1] == '1');
  connectNeon('#div-coinff',   (s) => s.div.ffs[2] == '1');
  connectNeon('#div-dpgamma',  (s) => s.div.ffs[3] == '1');
  connectNeon('#div-ngamma',   (s) => s.div.ffs[4] == '1');
  connectNeon('#div-psrcff',   (s) => s.div.ffs[5] == '1');
  connectNeon('#div-pringff',  (s) => s.div.ffs[6] == '1');
  connectNeon('#div-denomff',  (s) => s.div.ffs[7] == '1');
  connectNeon('#div-numrplus', (s) => s.div.ffs[8] == '1');
  connectNeon('#div-numrmin',  (s) => s.div.ffs[9] == '1');
  connectNeon('#div-qalpha',   (s) => s.div.ffs[10] == '1');
  connectNeon('#div-sac',      (s) => s.div.ffs[11] == '1');
  connectNeon('#div-m2',       (s) => s.div.ffs[12] == '1');
  connectNeon('#div-m1',       (s) => s.div.ffs[13] == '1');
  connectNeon('#div-nac',      (s) => s.div.ffs[14] == '1');
  connectNeon('#div-da',       (s) => s.div.ffs[15] == '1');
  connectNeon('#div-nalpha',   (s) => s.div.ffs[16] == '1');
  connectNeon('#div-dalpha',   (s) => s.div.ffs[17] == '1');
  connectNeon('#div-dgamma',   (s) => s.div.ffs[18] == '1');
  connectNeon('#div-npgamma',  (s) => s.div.ffs[19] == '1');
  connectNeon('#div-p2',       (s) => s.div.ffs[20] == '1');
  connectNeon('#div-p1',       (s) => s.div.ffs[21] == '1');
  connectNeon('#div-salpha',   (s) => s.div.ffs[22] == '1');
  connectNeon('#div-ds',       (s) => s.div.ffs[23] == '1');
  connectNeon('#div-nbeta',    (s) => s.div.ffs[24] == '1');
  connectNeon('#div-dbeta',    (s) => s.div.ffs[25] == '1');
  connectNeon('#div-ans1',     (s) => s.div.ffs[26] == '1');
  connectNeon('#div-ans2',     (s) => s.div.ffs[27] == '1');
  connectNeon('#div-ans3',     (s) => s.div.ffs[28] == '1');
  connectNeon('#div-ans4',     (s) => s.div.ffs[29] == '1');
  for (let i = 1; i <= 10; i++) {
    connectNeon(`#div-pring${i}`, ((i) => {
      return (s) => s.div.placeRing == i - 1;
    })(i));
  }

  makePanelSelectable('#div-front-panel');
  connectToggleSwitch('#div-heater-toggle');
  for (let i = 1; i <= 8; i++) {
    connectRotarySwitch(`#div-nr${i}`, [
      {value: 'α', degrees: -120},
      {value: 'β', degrees: -60},
      {value: '0', degrees: 0},
    ]);
    connectRotarySwitch(`#div-nc${i}`, [
      {value: 'C', degrees: 45},
      {value: '0', degrees: 0},
    ]);
    connectRotarySwitch(`#div-dr${i}`, [
      {value: 'α', degrees: -120},
      {value: 'β', degrees: -60},
      {value: '0', degrees: 0},
    ]);
    connectRotarySwitch(`#div-dc${i}`, [
      {value: 'C', degrees: 45},
      {value: '0', degrees: 0},
    ]);
    connectRotarySwitch(`#div-pl${i}`, [
      {value: 'D4', degrees: -180},
      {value: 'D7', degrees: -150},
      {value: 'D8', degrees: -115},
      {value: 'D9', degrees: -90},
      {value: 'D10', degrees: -55},
      {value: 'S4', degrees: -20},
      {value: 'S7', degrees: 0},
      {value: 'S8', degrees: 35},
      {value: 'S9', degrees: 65},
      {value: 'S10', degrees: 90},
    ]);
    connectRotarySwitch(`#div-ro${i}`, [
      {value: 'RO', degrees: 0},
      {value: 'NRO', degrees: -45},
    ]);
    connectRotarySwitch(`#div-an${i}`, [
      {value: '1', degrees: -130},
      {value: '2', degrees: -100},
      {value: '3', degrees: -70},
      {value: '4', degrees: -40},
      {value: 'OFF', degrees: 0},
    ]);
    connectRotarySwitch(`#div-il${i}`, [
      {value: 'I', degrees: 0},
      {value: 'NI', degrees: -30},
    ]);
  }

  makePanelSelectable('#div-bottom');
  for (let i = 1; i <= 8; i++) {
    connectNeon(`#div-prog${i}`, ((i) => {
      return (s) => s.div.program[i-1];
    })(i));
  }
  for (let i = 0; i < 9; i++) {
    connectNeon(`#div-progring${i}`, ((i) => {
      return (s) => s.div.progRing == i;
    })(i));
  }
}

function connectMultiplierElements(panelNumber) {
  const prefix = `mult${panelNumber}`;
  if (panelNumber == 1) {
    connectNeon('#mult1-reset', (s) => s.mult.reset1);
  } else if (panelNumber == 2) {
    for (let i = 1; i <= 14; i++) {
      connectNeon(`#mult2-stage${i}`, ((i) => {
        return (s) => s.mult.stage == i-1;
      })(i));
    }
  } else if (panelNumber == 3) {
    connectNeon('#mult3-reset', (s) => s.mult.reset3);
  }
  makePanelSelectable(`#${prefix}-front-panel`);
  connectToggleSwitch(`#${prefix}-heater-toggle`);
  const startDigit = [1, 9, 17][panelNumber - 1];
  for (let i = startDigit; i < startDigit + 8; i++) {
    const ierIcandSettings = [
      {value: 'α', degrees: -120},
      {value: 'β', degrees: -90},
      {value: 'γ', degrees: -60},
      {value: 'δ', degrees: -30},
      {value: 'ε', degrees: 0},
      {value: '0', degrees: 25},
    ];
    const ierIcandClearSettings = [
      {value: 'C', degrees: 45},
      {value: '0', degrees: 0},
    ];
    connectRotarySwitch(`#${prefix}-ieracc${i}`, ierIcandSettings);
    connectRotarySwitch(`#${prefix}-iercl${i}`, ierIcandClearSettings);
    connectRotarySwitch(`#${prefix}-icandacc${i}`, ierIcandSettings);
    connectRotarySwitch(`#${prefix}-icandcl${i}`, ierIcandClearSettings);
    connectRotarySwitch(`#${prefix}-sf${i}`, [
      {value: '10', degrees: -190},
      {value: '9', degrees: -155},
      {value: '8', degrees: -125},
      {value: '7', degrees: -95},
      {value: '6', degrees: -65},
      {value: '5', degrees: -35},
      {value: '4', degrees: 0},
      {value: '3', degrees: 25},
      {value: '2', degrees: 55},
      {value: '0', degrees: 85},
    ]);
    connectRotarySwitch(`#${prefix}-place${i}`, [
      {value: '2', degrees: -180},
      {value: '3', degrees: -150},
      {value: '4', degrees: -120},
      {value: '5', degrees: -90},
      {value: '6', degrees: -55},
      {value: '7', degrees: -25},
      {value: '8', degrees: 0},
      {value: '9', degrees: 30},
      {value: '10', degrees: 60},
    ]);
    connectRotarySwitch(`#${prefix}-prod${i}`, [
      {value: 'A', degrees: -160},
      {value: 'S', degrees: -130},
      {value: 'AS', degrees: -90},
      {value: '0', degrees: -60},
      {value: 'AC', degrees: -25},
      {value: 'SC', degrees: 0},
      {value: 'ASC', degrees: 30},
    ]);
    connectNeon(`#${prefix}-prog${i}`, ((i) => {
      return (s) => s.mult.program[i-1];
    })(i));
  }
}

function connectCT1Elements() {
  makePanelSelectable('#ct1-top');
  makePanelSelectable('#ct1-front-panel');
  makePanelSelectable('#ct1-bottom');
  for (let i = 1; i <= 30; i++) {
    connectNeon(`#ct1-prog${i}`, ((i) => {
      return (s) => s.constant[i-1] == '1';
    })(i));
  }
  
  connectToggleSwitch('#ct1-heater-toggle');
  const constantNames = 'abcdefghjk';
  for (let i = 1; i <= 30; i++) {
    const k = ~~((i - 1) / 6) * 2;
    const [c1, c2] = [constantNames[k], constantNames[k+1]];
    connectRotarySwitch(`#ct1-s${i}`, [
      {value: `${c1}l`, degrees: -120},
      {value: `${c1}r`, degrees: -90},
      {value: `${c1}lr`, degrees: -60},
      {value: `${c2}l`, degrees: -20},
      {value: `${c2}r`, degrees: 0},
      {value: `${c2}lr`, degrees: 25},
    ]);
  }
}

function connectCT2Elements() {
  makePanelSelectable('#ct2-front-panel');
  connectToggleSwitch('#ct2-heater-toggle');
  const pmSettings = [
    {value: 'P', degrees: 0},
    {value: 'M', degrees: -60},
  ];
  connectRotarySwitch('#ct2-jl', pmSettings);
  connectRotarySwitch('#ct2-jr', pmSettings);
  connectRotarySwitch('#ct2-kl', pmSettings);
  connectRotarySwitch('#ct2-kr', pmSettings);
  const digitSettings = [
    {value: '0', degrees: -175},
    {value: '1', degrees: -150},
    {value: '2', degrees: -120},
    {value: '3', degrees: -90},
    {value: '4', degrees: -60},
    {value: '5', degrees: -30},
    {value: '6', degrees: 0},
    {value: '7', degrees: 25},
    {value: '8', degrees: 55},
    {value: '9', degrees: 85},
  ];
  for (let i = 1; i <= 10; i++) {
    connectRotarySwitch(`#ct2-j${i}`, digitSettings);
    connectRotarySwitch(`#ct2-k${i}`, digitSettings);
  }
}

function connectPR1Elements() {
  makePanelSelectable('#pr1-front-panel');
  for (let i = 1; i <= 8; i++) {
    connectRotarySwitch(`#pr1-${i}-${i+1}`, [
      {value: '0', degrees: -60},
      {value: 'C', degrees: 0},
    ]);
  }
}

function connectPR2Elements() {
  makePanelSelectable('#pr2-front-panel');
  connectToggleSwitch('#pr2-heater-toggle');
  for (let i = 1; i <= 16; i++) {
    connectRotarySwitch(`#pr2-${i}`, [
      {value: '0', degrees: -60},
      {value: 'p', degrees: 0},
    ]);
  }
}

function connectPR3Elements() {
  makePanelSelectable('#pr3-front-panel');
  for (let i = 9; i <= 16; i++) {
    const n = i == 16 ? 1 : i + 1;
    connectRotarySwitch(`#pr3-${i}-${n}`, [
      {value: '0', degrees: -60},
      {value: 'C', degrees: 0},
    ]);
  }
}

window.onload = (event) => {
  const wrapper = document.querySelector('#eniac');
  const wrapperDoc = wrapper.contentDocument;
  addStylesToSvg(wrapperDoc);
  eniac = wrapperDoc.querySelector('#eniac');
  defaultViewBox = eniac.getAttribute('viewBox');
  connectInitiateElements();
  connectCyclingElements();
  connectMPElements(1);
  connectMPElements(2);
  for (let i = 1; i <= 3; i++) {
    connectFT1Elements(i);
    connectFT2Elements(i);
  }
  for (let i = 1; i <= 20; i++) {
    connectAccumulatorElements(i);
  }
  connectDividerElements();
  connectMultiplierElements(1);
  connectMultiplierElements(2);
  connectMultiplierElements(3);
  connectCT1Elements();
  connectCT2Elements();
  makePanelSelectable('#ct3-front-panel');
  connectPR1Elements();
  connectPR2Elements();
  connectPR3Elements();

  requestAnimationFrame(step);

  // for finding rotary switch settings
  const angle = document.querySelector('.angle');
  const angleValue = document.querySelector('.angle-value');
  angle.addEventListener('input', (event) => {
    if (adjustRotary) {
      adjustRotary(event.target.value);
      angleValue.textContent = event.target.value;
    }
  });
}

function viewElementSelector(name) {
  let elem = eniac.querySelector(name);
  let box = transformedBoundingBox(elem);
  eniac.setAttribute('viewBox', `${box.x} ${box.y} ${box.width} ${box.height}`);
}

function viewDefault() {
  eniac.setAttribute('viewBox', defaultViewBox);
}
