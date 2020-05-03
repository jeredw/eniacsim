let source = new EventSource('/events');
let eniac = null;
let defaultViewBox = '';
let neons = [];
let machineState = {};
let simulatorSwitches = {};

source.addEventListener('message', (event) => {
  machineState = JSON.parse(event.data);
  machineState["cycling"] = {
    "pulse": parseInt(machineState["cycling"], 10) / 2
  };
});

function step(ts) {
  for (const neon of neons) {
    neon(machineState);
  }
  requestAnimationFrame(step);
}

function extractState({unit,
                       unitIndex=undefined,
                       field=undefined,
                       fieldIndex=undefined,
                       eqValue=true}) {
  // s[unit][unitIndex][field][fieldIndex] == eqValue
  if (unit == "nil") {
    return (s) => false;
  }
  if (field === undefined) {
    return (s) => s[unit][unitIndex] == eqValue;
  }
  if (unitIndex === undefined) {
    return fieldIndex === undefined ?
      (s) => s[unit][field] == eqValue :
      (s) => s[unit][field][fieldIndex] == eqValue;
  }
  return fieldIndex === undefined ?
    (s) => s[unit][unitIndex][field] == eqValue :
    (s) => s[unit][unitIndex][field][fieldIndex] == eqValue;
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

function connectRotarySwitch(selector, simulatorName, settings, onChange=undefined) {
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
  const update = (newIndex) => {
    index = newIndex;
    const newValue = settings[index].value;
    if (onChange) {
      onChange(newValue);
    }
    rotation.setRotate(settings[index].degrees, cx, cy);
    // for manual adjustments
    adjustRotary = (degrees) => {
      rotation.setRotate(degrees, cx, cy);
    };
  };
  rotary.addEventListener('click', (event) => {
    event.stopPropagation();
    const delta = event.shiftKey ? -1 : 1;
    let newIndex = index + delta;
    if (newIndex >= settings.length) {
      newIndex = 0;
    }
    if (newIndex < 0) {
      newIndex = settings.length - 1;
    }
    update(newIndex);
    runCommands([`s ${simulatorName} ${settings[newIndex].value}`]);
  });
  if (simulatorName) {
    simulatorSwitches[simulatorName] = (value) => {
      const newIndex = settings.findIndex(s => s.value == value);
      if (newIndex != -1) {
        update(newIndex);
      }
    };
  }
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

function connectButton(selector, simulatorName) {
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
    if (simulatorName) {
      runCommands([`b ${simulatorName}`]);
    }
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
  const rotateNeedle1 = makeNeedleRotateable('#initiate-vm1-needle');
  const showRandomTrace = () => {
    const traces = eniac.querySelectorAll('.initiate-trace');
    const choice = ~~(Math.random() * traces.length);
    for (let i = 0; i < traces.length; i++) {
      traces[i].style.visibility = i == choice ? '' : 'hidden';
    }
    rotateNeedle1(Math.random() * 73);
  };
  connectRotarySwitch('#initiate-dcv1', '', [
    {value: 'A', degrees: -150},
    {value: 'B', degrees: -110},
    {value: 'C', degrees: -80},
    {value: 'D', degrees: -58},
    {value: 'E', degrees: -25},
    {value: 'F', degrees: 0},
    {value: 'G', degrees: 20},
    {value: 'H', degrees: 54},
  ], showRandomTrace);
  connectRotarySwitch('#initiate-dcv2', '', [
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
  connectRotarySwitch('#initiate-acv', '', [
    {value: 'VAB', degrees: -115},
    {value: 'VBC', degrees: -61},
    {value: 'VCA', degrees: 0},
  ], () => rotateNeedle2(Math.random() * 73));
}

function connectCyclingElements() {
  const showPulseTrace = (value) => {
    const traces = eniac.querySelectorAll('.cycling-trace');
    const selected = eniac.querySelector('#cycling-trace-' + value);
    for (const trace of traces) {
      trace.style.visibility = trace == selected ? '' : 'hidden';
    }
  };
  connectRotarySwitch('#cycling-osc', '', [
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
}

function configureSwitches(config) {
  for (const [selector, s] of Object.entries(config)) {
    switch (s.type) {
    case 'rotary':
      connectRotarySwitch(selector, s.simulatorName, s.settings);
      break;
    case 'toggle':
      connectToggleSwitch(selector);
      break;
    case 'button':
      connectButton(selector, s.simulatorName);
      break;
    }
  }
}

function configureNeons(config) {
  for (const [selector, predicate] of Object.entries(config)) {
    connectNeon(selector, extractState(predicate));
  }
}

function configurePanels(config) {
  for (const selector of config) {
    makePanelSelectable(selector);
  }
}

async function fetchConfig(file) {
  let response = await fetch(file);
  if (response.status != 200) {
    console.error(file, response.status);
    return;
  }
  return await response.json();
}

async function runCommands(commands) {
  let response = await fetch('/command', {
    method: 'post',
    headers: {
      'Content-type': 'application/json; charset=UTF-8',
    },
    body: JSON.stringify({"commands": commands})
  });
  if (response.status != 200) {
    console.error(response.status);
    return;
  }
  let data = await response.json();
  return data.outputs;
}

async function setSwitchesToSimulatorValues() {
  let switches = [];
  let commands = [];
  for (const switchName of Object.keys(simulatorSwitches)) {
    switches.push(switchName);
    commands.push(`s? ${switchName}`);
  }
  let outputs = await runCommands(commands);
  for (let i = 0; i < outputs.length; i++) {
    const switchName = switches[i];
    const value = outputs[i].trim();
    const update = simulatorSwitches[switchName];
    update(value.trim());
  }
}

window.onload = (event) => {
  const wrapper = document.querySelector('#eniac');
  const wrapperDoc = wrapper.contentDocument;
  addStylesToSvg(wrapperDoc);
  eniac = wrapperDoc.querySelector('#eniac');
  defaultViewBox = eniac.getAttribute('viewBox');

  fetchConfig('switches.json')
    .then(configureSwitches)
    .then(setSwitchesToSimulatorValues);
  fetchConfig('neons.json').then(configureNeons);
  fetchConfig('panels.json').then(configurePanels);
  connectInitiateElements();
  connectCyclingElements();

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

function getTransformToElement(elem, toElement) {
  return toElement.getScreenCTM().inverse().multiply(elem.getScreenCTM());
}

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
