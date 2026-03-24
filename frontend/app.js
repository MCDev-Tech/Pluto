const searchBtn = document.querySelector('#searchBtn');
const searchStatus = document.querySelector('#searchStatus');
const sideStatus = document.querySelector('#sideStatus');
const resultList = document.querySelector('#resultList');

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

async function fetchSourceText(version, mappingType, classPath) {
  const toUrl = (path) => {
    const url = new URL(path, window.location.origin);
    url.searchParams.set('version', version);
    url.searchParams.set('type', mappingType);
    url.searchParams.set('class', classPath);
    return url;
  };
  const getSource = async () => fetch(toUrl('/api/source/get'));

  let response = await getSource();
  if (response.ok) return await response.text();

  if (response.status !== 404 && response.status !== 412) {
    throw new Error(`${response.status} ${await response.text()}`);
  }

  const decompileResp = await fetch(toUrl('/api/source/decompile'));
  if (!decompileResp.ok && decompileResp.status !== 202) {
    throw new Error(`decompile failed: ${decompileResp.status} ${await decompileResp.text()}`);
  }

  let retries = 15;
  while (retries-- > 0) {
    await new Promise((r) => setTimeout(r, 2200));
    response = await getSource();
    if (response.ok) {
      return await response.text();
    }
    if (response.status !== 404 && response.status !== 412) {
      throw new Error(`${response.status} ${await response.text()}`);
    }
  }
  throw new Error('超时未生成源码，请稍后重试');
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
    <h3>${mainName} <span class="badge-type">${mainType}</span></h3>
    <div class="row"><div class="key">Named</div><div class="value">${named.Name || '-'}</div><button class="copy-btn">复制 Named</button></div>
    <div class="row"><div class="key">Notch</div><div class="value">${notch.Name || '-'}</div><button class="copy-btn">复制 Notch</button></div>
    <div class="row"><div class="key">类</div><div class="value">${mainClass || '-'}</div><button class="copy-btn">复制 类</button></div>
    <div class="row"><div class="key">签名</div><div class="value">${signature || '-'}</div><button class="copy-btn">复制 签名</button></div>
    <div class="row"><div class="key">AW</div><div class="value">${notch.Name || '-'}</div><button class="copy-btn">复制 AW</button></div>
    <div class="row"><div class="key">AT</div><div class="value">${named.Name || '-'}</div><button class="copy-btn">复制 AT</button></div>
    <div class="row"><button class="copy-btn">复制 翻译</button><button class="source-btn">查看源码</button></div>
    <div class="source-expanded hidden"><pre>未加载</pre></div>
  `;

  const [copyNamed, copyNotch, copyClass, copySignature, copyAw, copyAt, copyTranslated] = card.querySelectorAll('.copy-btn');
  copyNamed.addEventListener('click', () => copyText(named.Name || '', 'Named'));
  copyNotch.addEventListener('click', () => copyText(notch.Name || '', 'Notch'));
  copyClass.addEventListener('click', () => copyText(mainClass, '类'));
  copySignature.addEventListener('click', () => copyText(signature, '签名'));
  copyAw.addEventListener('click', () => copyText(notch.Name || '', 'AW'));
  copyAt.addEventListener('click', () => copyText(named.Name || '', 'AT'));
  copyTranslated.addEventListener('click', () => copyText(translated.Name || '', '翻译'));

  const sourceBtn = card.querySelector('.source-btn');
  const sourceBlock = card.querySelector('.source-expanded');
  const sourcePre = sourceBlock.querySelector('pre');

  sourceBtn.addEventListener('click', async () => {
    if (!classPath) {
      sourcePre.textContent = '无可用类路径';
      sourceBlock.classList.remove('hidden');
      return;
    }

    if (!sourceBlock.classList.contains('hidden')) {
      sourceBlock.classList.add('hidden');
      return;
    }

    sourceBlock.classList.remove('hidden');
    sourcePre.textContent = '加载中...';
    try {
      const sourceText = await fetchSourceText(version, mappingType, classPath);
      sourcePre.textContent = sourceText || '源码为空';
    } catch (err) {
      sourcePre.textContent = `错误：${err.message}`;
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
    updateSideStatus('可点击查看源码在卡片下展开');
  } catch (err) {
    updateSearchStatus(`查询异常：${err.message}`, true);
  }
}

searchBtn.addEventListener('click', searchMapping);
