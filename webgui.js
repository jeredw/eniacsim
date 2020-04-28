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

source.addEventListener('message', (event) => {
  //console.log(e.data);
});

function addStylesToSvg(doc) {
  let style = document.createElement('style');
  doc.documentElement.appendChild(style);
  style.sheet.insertRule('svg text { user-select: none; }');
}

let adjustRotary = null;

function connectRotarySwitch(selector, settings, onChange=undefined) {
  const rotary = eniac.querySelector(selector);
  const svg = rotary.ownerSVGElement;
  const wiper = rotary.querySelector('path');
  const box = wiper.getBBox();
  const [cx, cy] = [box.x + box.width / 2, box.y + box.height / 2];
  rotary.style.contain = 'layout paint';
  wiper.style.transition = 'transform 100ms linear';
  let rotation = svg.createSVGTransform();
  rotation.setRotate(0, cx, cy);
  wiper.transform.baseVal.appendItem(rotation);
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

function connectInitiateElements() {
  const panel = eniac.querySelector('#initiate-panel');
  panel.addEventListener('click', (event) => {
    viewElementSelector('#initiate-panel');
  });
  panel.addEventListener('contextmenu', (event) => {
    event.preventDefault();
    viewDefault();
  });
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
}

function connectCyclingElements() {
  const panel = eniac.querySelector('#cycling-panel');
  panel.addEventListener('click', (event) => {
    viewElementSelector('#cycling-panel');
  });
  panel.addEventListener('contextmenu', (event) => {
    event.preventDefault();
    viewDefault();
  });
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
  connectToggleSwitch('#cycling-heater-toggle');
  connectButton('#cycling-p');
}

function connectMP1Elements() {
  const panel = eniac.querySelector('#mp1-panel');
  panel.addEventListener('click', (event) => {
    viewElementSelector('#mp1-panel');
  });
  panel.addEventListener('contextmenu', (event) => {
    event.preventDefault();
    viewDefault();
  });
  connectToggleSwitch('#mp1-heater-toggle');
  connectRotarySwitch('#mp1-assoc-ab', [
    {value: 'a', degrees: -35},
    {value: 'b', degrees: 0},
  ]);
  connectRotarySwitch('#mp1-assoc-bc', [
    {value: 'b', degrees: -35},
    {value: 'c', degrees: 0},
  ]);
  connectRotarySwitch('#mp1-assoc-cd', [
    {value: 'c', degrees: -35},
    {value: 'd', degrees: 0},
  ]);
  connectRotarySwitch('#mp1-assoc-de', [
    {value: 'd', degrees: -35},
    {value: 'e', degrees: 0},
  ]);
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
  for (let decade = 20; decade >= 11; decade--) {
    for (let digit = 1; digit <= 6; digit++) {
      const name = `d${decade}s${digit}`;
      connectRotarySwitch(`#mp1-${name}`, decadeSettings);
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
  connectRotarySwitch(`#mp1-ca`, clearSettings);
  connectRotarySwitch(`#mp1-cb`, clearSettings);
  connectRotarySwitch(`#mp1-cc`, clearSettings);
  connectRotarySwitch(`#mp1-cd`, clearSettings);
  connectRotarySwitch(`#mp1-ce`, clearSettings);
}

window.onload = (event) => {
  const wrapper = document.querySelector('#eniac');
  const wrapperDoc = wrapper.contentDocument;
  addStylesToSvg(wrapperDoc);
  eniac = wrapperDoc.querySelector('#eniac');
  defaultViewBox = eniac.getAttribute('viewBox');
  connectInitiateElements();
  connectCyclingElements();
  connectMP1Elements();

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
