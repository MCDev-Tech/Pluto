const searchBtn = document.querySelector('#searchBtn');
const status = document.querySelector('#searchStatus');
const sideStatus = document.querySelector('#sideStatus');
const resultBody = document.querySelector('#resultTable tbody');
const sourceResult = document.querySelector('#sourceResult code');

window.addEventListener('load', async _ => {
  document.getElementById('api-version').innerText = `Backend Version: ${await fetch('/api/version').then(res => res.json()).then(json => json.version ?? 'Unknown')}`
})

function updateStatus(msg, isError = false) {
  status.textContent = msg;
  status.style.color = isError ? '#ff8a9a' : '#b7d4ff';
}

function updateSideStatus(msg, isError = false) {
  sideStatus.textContent = msg;
  sideStatus.style.color = isError ? '#ff9898' : '#9acdf6';
}

function buildRow(item, index, version, mappingType, translateType) {
  const tr = document.createElement('tr');
  tr.innerHTML = `
    <td>${index + 1}</td>
    <td>${item.notch.Type || item.named.Type || item.translated?.Type || 'unknown'}</td>
    <td title="Notch: ${item.notch.Name}">${item.notch.Name}</td>
    <td title="Named: ${item.named.Name}">${item.named.Name}</td>
    <td>${item.notch.Class || item.Named.Class || ''}</td>
    <td>${item.notch.Signature || item.Named.Signature || ''}</td>
    <td><button class="small-btn" data-class="${encodeURIComponent(item.notch.Class || item.named.Class || '')}" data-type="${item.notch.Type || item.named.Type || ''}">查看源码</button></td>
  `;

  const viewBtn = tr.querySelector('button');
  viewBtn.addEventListener('click', () => {
    const className = decodeURIComponent(viewBtn.dataset.class);
    if (!className) {
      updateSideStatus('没有可用类路径，用 Notch.Class 或 Named.Class', true);
      return;
    }
    loadSource(version, mappingType, className);
  });

  return tr;
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
    updateStatus('版本、类型与关键字（至少 3 字）为必填'.toString(), true);
    return;
  }

  updateStatus('正在查询，请稍候...');
  sideStatus.textContent = '展示源代码进度请在结果中点击。';
  resultBody.innerHTML = '';

  try {
    const url = new URL('/api/mapping/search', window.location.origin);
    url.searchParams.set('version', version);
    url.searchParams.set('type', mappingType);
    url.searchParams.set('keyword', keyword);
    if (translateType) {
      url.searchParams.set('translate', translateType);
    }

    const response = await fetch(url.toString());
    if (!response.ok) {
      const text = await response.text();
      updateStatus(`查询失败：${response.status} ${text}`, true);
      return;
    }

    const results = await response.json();
    const filtered = results.filter(item => {
      const t = (item.Notch && item.Notch.Type) || (item.Named && item.Named.Type) || '';
      if (t === 'class' && !showClass) return false;
      if (t === 'method' && !showMethod) return false;
      if (t === 'field' && !showField) return false;
      return true;
    });

    if (!filtered.length) {
      updateStatus('未找到符合筛选条件的结果', false);
      return;
    }

    filtered.forEach((item, index) => {
      resultBody.appendChild(buildRow(item, index, version, mappingType, translateType));
    });
    updateStatus(`已找到 ${filtered.length} 条结果`);
  } catch (err) {
    updateStatus(`查询异常：${err.message}`, true);
  }
}

async function loadSource(version, mappingType, clazz) {
  sourceResult.textContent = '';
  updateSideStatus(`请求源代码: ${clazz}`);

  const fetchSource = async () => {
    const url = new URL('/api/source/get', window.location.origin);
    url.searchParams.set('version', version);
    url.searchParams.set('type', mappingType);
    url.searchParams.set('class', clazz);
    return fetch(url.toString());
  };

  const sleep = ms => new Promise(resolve => setTimeout(resolve, ms));

  let resp = await fetchSource();

  if (resp.status === 412 || resp.status === 404) {
    updateSideStatus('源代码尚未准备，触发编译并等待结果...');

    const decompileUrl = new URL('/api/source/decompile', window.location.origin);
    decompileUrl.searchParams.set('version', version);
    decompileUrl.searchParams.set('type', mappingType);
    const decompileResp = await fetch(decompileUrl.toString());

    if (!decompileResp.ok && decompileResp.status !== 202) {
      const text = await decompileResp.text();
      updateSideStatus(`decompile 失败: ${decompileResp.status} ${text}`, true);
      return;
    }

    let maxRetry = 15;
    while (maxRetry > 0) {
      await sleep(2200);
      const polling = await fetchSource();
      if (polling.status === 200) {
        resp = polling;
        break;
      }
      if (polling.status === 404 || polling.status === 412) {
        maxRetry -= 1;
        updateSideStatus(`源代码准备中，剩余重试 ${maxRetry}`);
        continue;
      }
      const text = await polling.text();
      updateSideStatus(`获取失败：${polling.status} ${text}`, true);
      return;
    }

    if (maxRetry <= 0) {
      updateSideStatus('源代码获取超时，请稍后再试。', true);
      return;
    }
  }

  if (!resp.ok) {
    const text = await resp.text();
    updateSideStatus(`获取源代码失败：${resp.status} ${text}`, true);
    return;
  }

  const code = await resp.text();
  sourceResult.textContent = code || '源代码为空。';
  updateSideStatus('源代码加载完成。');
}

searchBtn.addEventListener('click', searchMapping);
