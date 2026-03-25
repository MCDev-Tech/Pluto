window.onload = async _ => {
  document.getElementById('searchBtn').addEventListener('click', searchMapping)
  document.getElementById('keyword').addEventListener('keydown', e => {
    if (e.key == 'Enter') this.searchMapping()
  })
  document.getElementById('download').addEventListener('click', e => window.open('https://github.com/MCDev-Tech/Pluto/releases', '_blank'))
  document.getElementById('api-version').innerHTML = 'Backend v' + await fetch('/api').then(res => res.json()).then(json => json.version)
  loadMCVersions()
}

async function loadMCVersions() {
  let versions = await fetch('versions.json').then(res => res.json()).catch(err => console.log(err))
  const versionSelect = document.getElementById('version')
  versionSelect.innerHTML = ''
  for (let version of versions) {
    let option = document.createElement('option')
    option.value = version
    option.innerText = version
    versionSelect.appendChild(option)
  }
  versionSelect.value = '1.20.1'
}

function updateSearchStatus(msg, isError = false) {
  const searchStatus = document.getElementById('searchStatus');
  searchStatus.textContent = msg;
  searchStatus.style.color = isError ? '#ff5d73' : '#8fb2d5';
}

function toSlashClass(className) {
  if (!className) return '';
  return className.includes('/') ? className : className.replace(/\./g, '/');
}

async function copyText(value, textDiv) {
  if (!value) return;
  await navigator.clipboard.writeText(value).catch(err => console.log(err));
  textDiv.innerText = '已复制！'
  setTimeout(() => textDiv.innerText = '点击复制', 500)
}

function getSourceUrl(version, mappingType, classPath) {
  const url = new URL('/api/source/get', window.location.origin);
  url.searchParams.set('version', version);
  url.searchParams.set('type', mappingType);
  url.searchParams.set('class', classPath);
  return url;
}

async function fetchSourceText(version, mappingType, classPath) {
  const getSource = async () => fetch(getSourceUrl(version, mappingType, classPath));

  let response = await getSource();
  if (response.ok) return await response.text();
  throw new Error(`${response.status} ${await response.text()}`);
}

function getLast(string, separator) {
  return string.split(separator).reverse()[0]
}

function buildName(data) {
  switch (data.Type) {
    case 'class': return data.Name
    default: return getLast(data.Class, '.') + '.' + data.Name
  }
}

function buildAT(data) {
  switch (data.Type) {
    case 'class': return `public ${data.Class}`
    case 'method': return `public ${data.Class} ${data.Name}${data.Signature}`
    case 'field': return `public ${data.Class} ${data.Name}`
  }
  return 'Unknown Type'
}

function buildAW(data) {
  const classSignature = data.Signature.substring(1, data.Signature.length - 1)
  switch (data.Type) {
    case 'class': return `accessible class ${classSignature}`
    case 'method': return `accessible method ${data.Class.replaceAll('.', '/')} ${data.Name} ${data.Signature}`
    case 'field': return `accessible field ${classSignature} ${data.Name} ${data.Signature}`
  }
  return 'Unknown Type'
}

function renderResultCard(item, index, version, mappingType) {
  const named = item.Named || item.named || {};
  const notch = item.Notch || item.notch || {};
  const translated = item.Translated || item.translated || {}, hasTranslation = Object.keys(translated).length > 0

  const mainClass = named.Class || notch.Class || '';
  const mainType = (named.Type || 'unknown').toLowerCase();
  const signature = named.Signature || '';
  const classPath = toSlashClass(mainClass);
  const namedAT = buildAT(named), namedAW = buildAW(named)

  function createSpan(text, clazz) {
    let span = document.createElement('span');
    span.innerText = text;
    if (clazz) span.className = clazz;
    return span;
  }
  function createCopyButton(content, big = false) {
    let main = document.createElement('div');
    main.className = 'tooltip-container';
    let image = document.createElement('img');
    image.className = big ? 'copy-btn-big' : 'copy-btn';
    image.src = 'copy-icon.svg';
    main.appendChild(image);
    let div = document.createElement('div');
    div.className = 'tooltip-text';
    div.innerText = '点击复制';
    main.appendChild(div);
    image.addEventListener('click', () => copyText(content, div));
    return main;
  }
  function createSourceButton() {
    let main = document.createElement('div');
    main.className = 'source-btn tooltip-container';
    let image = document.createElement('img');
    image.className = 'source-btn'
    image.src = 'source-icon.svg';
    main.appendChild(image);
    let div = document.createElement('div');
    div.className = 'tooltip-text';
    div.innerText = '查看源代码';
    main.appendChild(div);
    return main;
  }
  const card = document.createElement('article');
  card.className = 'result-card';

  const title = document.createElement('h3');
  title.appendChild(createSpan(named.Name));
  title.appendChild(createCopyButton(named.Name, true));
  if (hasTranslation) {
    title.appendChild(createSpan('>>', 'key'))
    title.appendChild(createSpan(' '))
    title.appendChild(createSpan(translated.Name));
    title.appendChild(createCopyButton(translated.Name, true))
  }
  title.appendChild(createSpan(mainType, 'badge-type badge-type-' + mainType));
  const sourceButton = title.appendChild(createSourceButton());
  card.appendChild(title);

  const divTranslate = document.createElement('div')
  divTranslate.className = 'row'
  if (hasTranslation) {
    divTranslate.appendChild(createSpan(buildName(named), 'value'))
    divTranslate.appendChild(createSpan(' '))
    divTranslate.appendChild(createCopyButton(buildName(named)))
    divTranslate.appendChild(createSpan('>', 'key'))
    divTranslate.appendChild(createSpan(' '))
  }
  divTranslate.appendChild(createSpan(buildName(notch), 'value'))
  divTranslate.appendChild(createSpan(' '))
  divTranslate.appendChild(createCopyButton(buildName(notch)))
  divTranslate.appendChild(createSpan('>', 'key'))
  divTranslate.appendChild(createSpan(' '))
  divTranslate.appendChild(createSpan(buildName(hasTranslation ? translated : named), 'value'))
  divTranslate.appendChild(createSpan(' '))
  divTranslate.appendChild(createCopyButton(buildName(hasTranslation ? translated : named)))
  card.appendChild(divTranslate)
  card.appendChild(document.createElement('br'))

  const divSignature = document.createElement('div')
  divSignature.className = 'row'
  divSignature.appendChild(createSpan('签名：', 'key'))
  divSignature.appendChild(createSpan(signature, 'value'))
  divSignature.appendChild(createSpan(' '))
  divSignature.appendChild(createCopyButton(signature))
  card.appendChild(divSignature)
  card.appendChild(document.createElement('br'))

  card.appendChild(document.createElement('hr'))

  const divAT = document.createElement('div')
  divAT.className = 'row'
  divAT.appendChild(createSpan('AT：', 'key'))
  divAT.appendChild(createSpan(namedAT, 'value'))
  divAT.appendChild(createSpan(' '))
  divAT.appendChild(createCopyButton(namedAT))
  card.appendChild(divAT)
  card.appendChild(document.createElement('br'))

  const divAW = document.createElement('div')
  divAW.className = 'row'
  divAW.appendChild(createSpan('AW：', 'key'))
  divAW.appendChild(createSpan(namedAW, 'value'))
  divAW.appendChild(createSpan(' '))
  divAW.appendChild(createCopyButton(namedAW))
  card.appendChild(divAW)

  const divSource = document.createElement('div')
  divSource.className = 'source-expanded hidden'
  const pre = document.createElement('pre')
  pre.className = 'line-numbers'
  pre['data-src'] = getSourceUrl(version, mappingType, classPath)
  pre['data-download-link'] = true
  const code = document.createElement('code')
  code.className = 'language-java'
  code.innerText = '未加载'
  pre.appendChild(code)
  divSource.append(pre)
  card.append(divSource)

  sourceButton.addEventListener('click', async () => {
    if (!classPath) {
      code.textContent = '无可用类路径';
      divSource.classList.remove('hidden');
      return;
    }

    if (!divSource.classList.contains('hidden')) {
      divSource.classList.add('hidden');
      return;
    }

    divSource.classList.remove('hidden');
    code.textContent = '加载中...';
    try {
      const sourceText = await fetchSourceText(version, mappingType, classPath);
      code.textContent = sourceText || '源码为空';
      //Apply highlight
      Prism.highlightAll();
    } catch (err) {
      code.textContent = `错误：${err.message}`;
    }
  });

  return card;
}

async function searchMapping() {
  const resultList = document.getElementById('resultList');

  const version = document.getElementById('version').value.trim();
  const mappingType = document.getElementById('mappingType').value;
  const translateType = document.getElementById('translateType').value;
  const keyword = document.getElementById('keyword').value.trim();
  const showClass = document.getElementById('filterClass').checked;
  const showMethod = document.getElementById('filterMethod').checked;
  const showField = document.getElementById('filterField').checked;

  if (!version || !mappingType || keyword.length < 3) {
    updateSearchStatus('MC版本、命名空间 和 关键字（最少3字符）为必填', true);
    return;
  }

  updateSearchStatus('正在查询...');
  resultList.innerHTML = '';

  try {
    const url = new URL('/api/mapping/search', window.location.origin);
    url.searchParams.set('version', version);
    url.searchParams.set('type', mappingType);
    url.searchParams.set('keyword', keyword);
    url.searchParams.set('filter', (showClass << 2) + (showMethod << 1) + (showField << 0));
    if (translateType) {
      url.searchParams.set('translate', translateType);
    }

    const response = await fetch(url);
    if (!response.ok) {
      const text = await response.text();
      updateSearchStatus(`查询失败 ${response.status}: ${text}`, true);
      return;
    }

    const results = await response.json();
    const filtered = results.filter((item) => {
      const t = item.named.Type;
      if (t === 'class' && !showClass) return false;
      if (t === 'method' && !showMethod) return false;
      if (t === 'field' && !showField) return false;
      return true;
    });

    if (!filtered.length) {
      updateSearchStatus('未找到结果');
      return;
    }

    filtered.forEach((item, index) => {
      resultList.appendChild(renderResultCard(item, index, version, mappingType));
    });

    updateSearchStatus(`共 ${filtered.length} 条结果`);
  } catch (err) {
    updateSearchStatus(`查询异常：${err.message}`, true);
  }
}
