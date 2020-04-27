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

function connectRotarySwitch(selector, settings) {
  const rotary = eniac.querySelector(selector)
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
    rotary.dataset.value = settings[index].value;
    rotation.setRotate(settings[index].degrees, cx, cy);
  });
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
  connectRotarySwitch('#initiate-dcv1', [
    {value: 'A', degrees: -150},
    {value: 'B', degrees: -110},
    {value: 'C', degrees: -80},
    {value: 'D', degrees: -58},
    {value: 'E', degrees: -25},
    {value: 'F', degrees: 0},
    {value: 'G', degrees: 20},
    {value: 'H', degrees: 54},
  ]);
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
  ]);
  connectRotarySwitch('#initiate-acv', [
    {value: 'VAB', degrees: -115},
    {value: 'VBC', degrees: -61},
    {value: 'VCA', degrees: 0},
  ]);
}

window.onload = (event) => {
  let wrapper = document.querySelector('#eniac');
  let wrapperDoc = wrapper.contentDocument;
  addStylesToSvg(wrapperDoc);
  eniac = wrapperDoc.querySelector('#eniac');
  defaultViewBox = eniac.getAttribute('viewBox');
  connectInitiateElements();
};

function viewElementSelector(name) {
  let elem = eniac.querySelector(name);
  let box = transformedBoundingBox(elem);
  eniac.setAttribute('viewBox', box.x + ' ' + box.y + ' ' + box.width + ' ' + box.height);
}

function viewDefault() {
  eniac.setAttribute('viewBox', defaultViewBox);
}
