const searchBtn = document.getElementById('searchBtn');
const searchStatus = document.getElementById('searchStatus');
const sideStatus = document.getElementById('sideStatus');
const resultList = document.getElementById('resultList');

function updateSearchStatus(msg, isError = false) {
  searchStatus.textContent = msg;
  searchStatus.style.color = isError ? '#ff5d73' : '#8fb2d5';
}

function updateSideStatus(msg, isError = false) {
  sideStatus.textContent = msg;
  sideStatus.style.color = isError ? '#ff5d73' : '#8fb2d5';
}

function toSlashClass(className) {
  if (!className) return '';
  return className.includes('/') ? className : className.replace(/\./g, '/');
}

async function copyText(value, label) {
  if (!value) return;
  try {
    await navigator.clipboard.writeText(value);
    updateSideStatus(`已复制 ${label}`);
  } catch (e) {
    updateSideStatus(`复制失败: ${e.message}`, true);
  }
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

function renderResultCard(item, index, version, mappingType) {
  const named = item.Named || item.named || {};
  const notch = item.Notch || item.notch || {};
  const translated = item.Translated || item.translated || {};
  const mainName = named.Name || notch.Name || '';
  const mainClass = named.Class || notch.Class || '';
  const mainType = (named.Type || notch.Type || translated.Type || 'unknown').toLowerCase();
  const signature = named.Signature || notch.Signature || '';
  const classPath = toSlashClass(mainClass);

  const card = document.createElement('article');
  card.className = 'result-card';
  card.innerHTML = `
    <h3><span>${mainName}</span><image src="copy-icon.svg" class="copy-btn-big"></image><span class="badge-type">${mainType}</span><image src="source-icon.svg" class="source-btn"></image></h3>
    <div class="row">
      <span class="value">${notch.Name || '-'}</span>&nbsp;<image src="copy-icon.svg" class="copy-btn"></image>
      <span class="key">></span>
      <span class="value">${named.Name || '-'}</span>&nbsp;<image src="copy-icon.svg" class="copy-btn"></image>
    </div><br>
    <div class="row">
      <span class="key">签名：</span><span class="value">${signature || '-'}</span>&nbsp;<image src="copy-icon.svg" class="copy-btn">
    </div>
    <div class="source-expanded hidden"><pre class="line-numbers" data-src="${getSourceUrl(version, mappingType, classPath)}" data-download-link><code class="language-java">未加载</code></pre></div>
  `;

  const [copyMainName] = card.querySelectorAll('.copy-btn-big');
  const [copyNotch, copyNamed, copySignature] = card.querySelectorAll('.copy-btn');
  copyMainName.addEventListener('click', () => copyText(mainName || '', 'Named'));
  copyNotch.addEventListener('click', () => copyText(notch.Name || '', 'Notch'));
  copyNamed.addEventListener('click', () => copyText(named.Name || '', 'Named'));
  copySignature.addEventListener('click', () => copyText(signature, '签名'));

  const sourceBtn = card.querySelector('.source-btn');
  const sourceBlock = card.querySelector('.source-expanded');
  const sourceCode = sourceBlock.querySelector('code');

  sourceBtn.addEventListener('click', async () => {
    if (!classPath) {
      sourceCode.textContent = '无可用类路径';
      sourceBlock.classList.remove('hidden');
      return;
    }

    if (!sourceBlock.classList.contains('hidden')) {
      sourceBlock.classList.add('hidden');
      return;
    }

    sourceBlock.classList.remove('hidden');
    sourceCode.textContent = '加载中...';
    try {
      const sourceText = await fetchSourceText(version, mappingType, classPath);
      sourceCode.textContent = sourceText || '源码为空';
      //Apply highlight
      Prism.highlightAll();
    } catch (err) {
      sourceCode.textContent = `错误：${err.message}`;
    }
  });

  return card;
}

async function searchMapping() {
  const version = document.querySelector('#version').value.trim();
  const mappingType = document.querySelector('#mappingType').value;
  const translateType = document.querySelector('#translateType').value;
  const keyword = document.querySelector('#keyword').value.trim();
  const showClass = document.querySelector('#filterClass').checked;
  const showMethod = document.querySelector('#filterMethod').checked;
  const showField = document.querySelector('#filterField').checked;

  if (!version || !mappingType || keyword.length < 3) {
    updateSearchStatus('MC版本、命名空间 和 关键字（最少3字符）为必填', true);
    return;
  }

  updateSearchStatus('正在查询...');
  updateSideStatus('请求中...');
  resultList.innerHTML = '';

  try {
    const url = new URL('/api/mapping/search', window.location.origin);
    url.searchParams.set('version', version);
    url.searchParams.set('type', mappingType);
    url.searchParams.set('keyword', keyword);
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
      const t = ((item.Named && item.Named.Type) || (item.Notch && item.Notch.Type) || '').toLowerCase();
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
    updateSideStatus('查询成功');
  } catch (err) {
    updateSearchStatus(`查询异常：${err.message}`, true);
  }
}

searchBtn.addEventListener('click', searchMapping);
