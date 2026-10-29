/* Same-origin WHOIS UI. Talks only to this deployment's /v1/whois and /multi. */
(function () {
  "use strict";

  var MAX_BATCH = 10;
  var RECENT_KEY = "whois-ui-recent";
  var LANG_KEY = "whois-ui-lang";

  var I18N = {
    zh: {
      title: "wi.fs.kg — 域名 WHOIS 查询",
      desc: "输入域名，一次看清注册状态、注册商、到期日、名称服务器和联系人。RDAP 优先，WHOIS 兜底。免注册、无广告、无追踪脚本。",
      "nav.lookup": "查询",
      "nav.record": "档案",
      "nav.ns": "服务器",
      "nav.batch": "批量",
      "nav.api": "接口",
      "nav.docs": "文档",
      "lang.aria": "切换到英文",
      "lang.btn": "EN",
      eyebrow: "WHOIS LOOKUP · 免注册 · 无广告",
      "h1.l1": "查一个域名，",
      "h1.a": "归属、",
      "h1.b": "状态、",
      "h1.c": "到期日",
      "h1.l3": "一次摊开。",
      "tag.main": "WHOIS Lookup",
      "tag.sub": "注册、注册商与到期，一次看清。",
      lede: "输入域名即可。优先查询 RDAP，这个后缀没有 RDAP 时再退回 WHOIS，两边都收成同一套字段。没有广告，没有追踪脚本，也不需要注册。命令行直接请求",
      "label.q": "域名",
      ph: "example.com",
      "btn.clear": "清空",
      "btn.search": "查询",
      "btn.searching": "查询中…",
      "chips.label": "试试",
      recent: "最近",
      "recent.clear": "清除",
      "status.idle": "输入域名，或点下面任一示例。",
      "status.loading": "正在向注册局查询…",
      "status.done": "已展开 {domain}",
      "bar.wait": "等待查询",
      "snap.title": "域名速览",
      "snap.badge.wait": "等待查询",
      "snap.badge.loading": "读取中",
      "snap.copy": "复制速览",
      "snap.fine": "字段来自本站 /v1/whois 的归一化结果。联系人常被注册局隐去，隐去时会标明，不用空值冒充「查无此人」。",
      "snap.sub.wait": "查询后这里摊开注册状态",
      "snap.f.status": "注册状态",
      "snap.f.registrar": "注册商",
      "snap.f.created": "创建日期",
      "snap.f.expires": "到期日期",
      "snap.f.source": "数据来源",
      "snap.f.tld": "顶级域",
      "record.title": "注册档案",
      "record.copy": "复制",
      "record.json": "JSON",
      "record.empty.title": "还没有查询目标",
      "record.empty": "输入一个域名，或点上方任意一个示例。",
      "record.domain": "域名",
      "record.unicode": "Unicode",
      "record.id": "注册局 ID",
      "record.tld": "顶级域",
      "record.registered": "注册状态",
      "record.registrar": "注册商",
      "record.iana": "IANA ID",
      "record.url": "注册商网址",
      "record.whois": "WHOIS 服务器",
      "record.reseller": "经销商",
      "record.abuseEmail": "滥用邮箱",
      "record.abusePhone": "滥用电话",
      "record.created": "创建",
      "record.updated": "更新",
      "record.expires": "到期",
      "record.statuses": "状态码",
      "record.statuses.hint": "EPP 状态，右侧是原文",
      "record.noStatus": "上游没有返回状态码",
      "ns.title": "名称服务器",
      "ns.empty": "查询后这里列出委派的名称服务器。",
      "ns.none": "这份结果里没有名称服务器。",
      "dnssec.wait": "查询后显示 DNSSEC",
      "dnssec.on": "DNSSEC 已签名",
      "dnssec.off": "DNSSEC 未签名",
      "dnssec.detail.on": "注册局返回了签名信息。",
      "dnssec.detail.off": "没有看到 DS 记录，或上游标明未签名。",
      "dnssec.ns": "名称服务器",
      "dnssec.ds": "DS 记录",
      "contacts.title": "联系人",
      "contacts.empty": "查询后按角色列出联系人。被隐去的会单独标明。",
      "contacts.redacted": "已隐去",
      "contacts.none": "没有公开字段",
      "role.registrant": "注册人",
      "role.admin": "管理联系人",
      "role.tech": "技术联系人",
      "role.billing": "账单联系人",
      "c.name": "姓名",
      "c.org": "组织",
      "c.email": "邮箱",
      "c.phone": "电话",
      "c.addr": "地址",
      "batch.title": "批量查询",
      "batch.help": "一次粘贴多个域名（换行、逗号或空格分隔均可），最多 10 个。这是服务端 /multi 的默认上限，公开部署上还有按 IP 的速率限制。",
      "batch.ph": "example.com\ngithub.com\nwikipedia.org",
      "batch.run": "解析并查询",
      "batch.running": "查询中…",
      "batch.csv": "导出 CSV",
      "batch.clear": "清空",
      "batch.empty": "结果表会在第一次批量查询后出现。",
      "batch.trim": "超过 10 个，只查了前 10 个。",
      "batch.need": "先贴几个域名。",
      "th.domain": "域名",
      "th.status": "状态",
      "th.registrar": "注册商",
      "th.expires": "到期",
      "th.source": "来源",
      "api.title": "终端接口",
      "api.badge": "无需密钥",
      "api.lede": "页面和接口在同一部署上。浏览器打开根路径看到这个界面；curl 打下面的路径拿到 JSON。不需要 API key，响应带 CORS。",
      "api.n1": "单条查询走 <code class=\"path\">/v1/whois/{domain}</code>，裸路径 <code class=\"path\">/{domain}</code> 也行。加上 <code class=\"path\">?raw=true</code> 返回上游 RDAP JSON 或 WHOIS 原文。",
      "api.n2": "批量用 <code class=\"path\">/multi?domains=a.com,b.com</code>。默认最多 10 个，可用环境变量 <code class=\"path\">MAX_DOMAINS</code> 调整。",
      "api.n3": "完整字段说明在 <code class=\"path\">/docs</code>。OpenAPI 规范在 <code class=\"path\">/openapi.yaml</code>。",
      "about.title": "数据与限制",
      "about.n1": "先问 RDAP。只有这个后缀没有 RDAP 服务器时，才退回端口 43 的 WHOIS。",
      "about.n2": "两种来源收成同一套字段：是否注册、注册商、日期、状态码、名称服务器、DNSSEC、联系人。",
      "about.n3": "速率看部署方式。放在 Vercel 上时，匿名访问默认大约每分钟 30 次。",
      "about.n4": "这是注册信息，不是「这个站安不安全」的结论。页面不加载广告或统计脚本。",
      "faq.title": "常见问题",
      "faq.q1": "怎么查一个域名的注册信息？",
      "faq.a1": "在上方输入域名，例如 example.com，或点一个示例。也接受带 https:// 的地址，服务端会先剥掉协议和路径。页面请求本站的 /v1/whois/{domain}，把注册状态、注册商、日期、名称服务器和联系人摊开。",
      "faq.q2": "为什么联系人经常是空的？",
      "faq.a2": "多数注册局按隐私规则隐去个人联系信息，通用顶级域尤其常见。隐去时这一格会写「已隐去」。那是注册局的选择，不是查询失败。",
      "faq.q3": "可以在命令行里查吗？",
      "faq.a3": "可以，不需要密钥。请求 <code class=\"path\">/v1/whois/example.com</code> 得到归一化 JSON；加上 <code class=\"path\">?raw=true</code> 则是上游原文。多个域名用 <code class=\"path\">/multi?domains=a.com,b.com</code>。完整地址在上面的终端示例里。浏览器访问根路径仍然是这个页面。",
      "faq.q4": "会记录我查过什么吗？",
      "faq.a4": "页面没有第三方统计。「最近」列表只存在你这台浏览器的本地存储里，可以在上面一键清除。服务端可以开内存缓存，用来少打注册局，缓存的是查询结果，不是一份用户档案。",
      "foot.blurb": "域名注册信息",
      "foot.privacy": "无广告 · 无追踪",
      "reg.yes": "已注册",
      "reg.no": "未注册",
      "source.rdap": "RDAP",
      "source.whois": "WHOIS",
      cached: "来自缓存",
      copied: "已复制",
      copyfail: "复制失败",
      "json.need": "还没有可以展开的结果",
      "json.hide": "收起",
      "err.EMPTY": "先输入一个域名",
      "err.INVALID_DOMAIN": "没能认出这是一个可注册的域名",
      "err.UNSUPPORTED_TLD": "这个后缀没有 RDAP，WHOIS 也没能查到",
      "err.RATE_LIMITED": "查询太频繁，稍后再试",
      "err.UPSTREAM_ERROR": "注册局暂时没有返回有效数据",
      "err.UPSTREAM_TIMEOUT": "注册局超时了",
      "err.INTERNAL_ERROR": "服务暂时出错",
      "err.UNAUTHORIZED": "这个部署开启了访问密钥",
      "err.NETWORK": "网络请求失败，请重试",
      "err.generic": "查询失败",
      "curl.c1": "归一化 JSON",
      "curl.c2": "上游原文",
      "curl.c3": "一次查多个",
      "sum.registered": "注册状态",
      "sum.registrar": "注册商",
      "sum.created": "创建",
      "sum.updated": "更新",
      "sum.expires": "到期",
      "sum.ns": "名称服务器",
      "sum.source": "来源",
      "sum.statuses": "状态码"
    },
    en: {
      title: "wi.fs.kg — Domain WHOIS lookup",
      desc: "Look up registration status, registrar, dates, nameservers, and contacts for a domain. RDAP first, WHOIS fallback. No account, no ads, no tracking.",
      "nav.lookup": "Lookup",
      "nav.record": "Record",
      "nav.ns": "Servers",
      "nav.batch": "Batch",
      "nav.api": "API",
      "nav.docs": "Docs",
      "lang.aria": "切换到中文",
      "lang.btn": "中文",
      eyebrow: "WHOIS LOOKUP · NO SIGNUP · NO ADS",
      "h1.l1": "Look up a domain.",
      "h1.a": "owner,",
      "h1.b": "status,",
      "h1.c": "expiry",
      "h1.l3": "in one view.",
      "tag.main": "WHOIS Lookup",
      "tag.sub": "Registrar, status, and expiry, opened up together.",
      lede: "Type a domain. RDAP is queried first; WHOIS on port 43 is the fallback when a suffix has no RDAP server. Both come back as the same fields. No ads, no tracking scripts, no account. From the shell, request",
      "label.q": "Domain",
      ph: "example.com",
      "btn.clear": "Clear",
      "btn.search": "Look up",
      "btn.searching": "Looking…",
      "chips.label": "Try",
      recent: "Recent",
      "recent.clear": "Clear",
      "status.idle": "Enter a domain, or pick an example.",
      "status.loading": "Asking the registry…",
      "status.done": "Opened {domain}",
      "bar.wait": "Waiting",
      "snap.title": "Domain brief",
      "snap.badge.wait": "Waiting",
      "snap.badge.loading": "Reading",
      "snap.copy": "Copy brief",
      "snap.fine": "Fields come from this site's normalized /v1/whois response. Contacts are often redacted by the registry; redaction is labeled, not shown as a failed lookup.",
      "snap.sub.wait": "Registration status lands here",
      "snap.f.status": "Status",
      "snap.f.registrar": "Registrar",
      "snap.f.created": "Created",
      "snap.f.expires": "Expires",
      "snap.f.source": "Source",
      "snap.f.tld": "TLD",
      "record.title": "Registration",
      "record.copy": "Copy",
      "record.json": "JSON",
      "record.empty.title": "Nothing looked up yet",
      "record.empty": "Enter a domain, or pick an example above.",
      "record.domain": "Domain",
      "record.unicode": "Unicode",
      "record.id": "Registry ID",
      "record.tld": "TLD",
      "record.registered": "Status",
      "record.registrar": "Registrar",
      "record.iana": "IANA ID",
      "record.url": "Registrar URL",
      "record.whois": "WHOIS server",
      "record.reseller": "Reseller",
      "record.abuseEmail": "Abuse email",
      "record.abusePhone": "Abuse phone",
      "record.created": "Created",
      "record.updated": "Updated",
      "record.expires": "Expires",
      "record.statuses": "Status codes",
      "record.statuses.hint": "EPP status, raw code on the right",
      "record.noStatus": "The upstream response had no status codes",
      "ns.title": "Nameservers",
      "ns.empty": "Delegated nameservers show up here after a lookup.",
      "ns.none": "This result did not include nameservers.",
      "dnssec.wait": "DNSSEC after a lookup",
      "dnssec.on": "DNSSEC signed",
      "dnssec.off": "DNSSEC not signed",
      "dnssec.detail.on": "The registry reported a signed delegation.",
      "dnssec.detail.off": "No DS records, or the upstream marked the zone unsigned.",
      "dnssec.ns": "Nameservers",
      "dnssec.ds": "DS records",
      "contacts.title": "Contacts",
      "contacts.empty": "Contacts are listed by role after a lookup. Redacted roles are labeled.",
      "contacts.redacted": "Redacted",
      "contacts.none": "No public fields",
      "role.registrant": "Registrant",
      "role.admin": "Admin",
      "role.tech": "Tech",
      "role.billing": "Billing",
      "c.name": "Name",
      "c.org": "Organization",
      "c.email": "Email",
      "c.phone": "Phone",
      "c.addr": "Address",
      "batch.title": "Batch lookup",
      "batch.help": "Paste domains separated by newlines, commas, or spaces. Up to 10, which is the server's default MAX_DOMAINS. Public deploys also rate-limit by IP.",
      "batch.ph": "example.com\ngithub.com\nwikipedia.org",
      "batch.run": "Resolve",
      "batch.running": "Looking…",
      "batch.csv": "Export CSV",
      "batch.clear": "Clear",
      "batch.empty": "The table fills in after the first batch lookup.",
      "batch.trim": "More than 10 domains; only the first 10 were queried.",
      "batch.need": "Paste a few domains first.",
      "th.domain": "Domain",
      "th.status": "Status",
      "th.registrar": "Registrar",
      "th.expires": "Expires",
      "th.source": "Source",
      "api.title": "Terminal API",
      "api.badge": "No key",
      "api.lede": "The page and the API are the same deployment. A browser at the root gets this UI; curl against the paths below gets JSON. No API key. CORS is open.",
      "api.n1": "One domain: <code class=\"path\">/v1/whois/{domain}</code>. The bare <code class=\"path\">/{domain}</code> path works too. <code class=\"path\">?raw=true</code> returns the upstream RDAP JSON or WHOIS text.",
      "api.n2": "Batch: <code class=\"path\">/multi?domains=a.com,b.com</code>. The default cap is 10, changeable with <code class=\"path\">MAX_DOMAINS</code>.",
      "api.n3": "Field reference lives at <code class=\"path\">/docs</code>. The OpenAPI document is <code class=\"path\">/openapi.yaml</code>.",
      "about.title": "Data & limits",
      "about.n1": "RDAP first. Port-43 WHOIS is only used when the suffix has no RDAP server.",
      "about.n2": "Both sources are normalized: registration, registrar, dates, status codes, nameservers, DNSSEC, contacts.",
      "about.n3": "Rate limits depend on the deploy. On Vercel, anonymous traffic defaults to about 30 requests a minute.",
      "about.n4": "This is registration data, not a verdict on whether a site is safe. The page loads no ads and no analytics.",
      "faq.title": "FAQ",
      "faq.q1": "How do I look up a domain?",
      "faq.a1": "Type a domain such as example.com, or pick an example. URLs are fine — the server strips the scheme and path. The page calls this site's /v1/whois/{domain} and lays out status, registrar, dates, nameservers, and contacts.",
      "faq.q2": "Why are contacts often empty?",
      "faq.a2": "Most registries redact personal contact data, especially on generic TLDs. A redacted role is labeled. That is the registry's choice, not a failed lookup.",
      "faq.q3": "Can I use this from the shell?",
      "faq.a3": "Yes, with no key. <code class=\"path\">/v1/whois/example.com</code> returns normalized JSON; <code class=\"path\">?raw=true</code> returns the upstream payload. Several domains: <code class=\"path\">/multi?domains=a.com,b.com</code>. The terminal block above shows the absolute URL for this host. A browser at the root still gets this page.",
      "faq.q4": "Do you store what I look up?",
      "faq.a4": "There is no third-party analytics. The recent list lives only in this browser and can be cleared above. The server may keep an in-memory cache of lookup results so it hits registries less often. That cache is not a user profile.",
      "foot.blurb": "Domain registration",
      "foot.privacy": "No ads · no tracking",
      "reg.yes": "Registered",
      "reg.no": "Not registered",
      "source.rdap": "RDAP",
      "source.whois": "WHOIS",
      cached: "Cached",
      copied: "Copied",
      copyfail: "Could not copy",
      "json.need": "Nothing to expand yet",
      "json.hide": "Hide",
      "err.EMPTY": "Enter a domain first",
      "err.INVALID_DOMAIN": "That is not a registrable domain",
      "err.UNSUPPORTED_TLD": "No RDAP server for this suffix, and WHOIS did not answer",
      "err.RATE_LIMITED": "Too many lookups. Try again shortly.",
      "err.UPSTREAM_ERROR": "The registry did not return a usable response",
      "err.UPSTREAM_TIMEOUT": "The registry timed out",
      "err.INTERNAL_ERROR": "The service hit an error",
      "err.UNAUTHORIZED": "This deploy requires an API key",
      "err.NETWORK": "The request failed. Try again.",
      "err.generic": "Lookup failed",
      "curl.c1": "Normalized JSON",
      "curl.c2": "Upstream payload",
      "curl.c3": "Several domains",
      "sum.registered": "Status",
      "sum.registrar": "Registrar",
      "sum.created": "Created",
      "sum.updated": "Updated",
      "sum.expires": "Expires",
      "sum.ns": "Nameservers",
      "sum.source": "Source",
      "sum.statuses": "Status codes"
    }
  };

  var STATUS_GLOSS = {
    ok: { zh: "正常", en: "OK" },
    inactive: { zh: "未激活", en: "Inactive" },
    clientTransferProhibited: { zh: "禁止转移", en: "Transfer locked" },
    clientDeleteProhibited: { zh: "禁止删除", en: "Delete locked" },
    clientUpdateProhibited: { zh: "禁止更新", en: "Update locked" },
    clientRenewProhibited: { zh: "禁止续费", en: "Renew locked" },
    clientHold: { zh: "注册商暂停", en: "Client hold" },
    serverTransferProhibited: { zh: "注册局禁止转移", en: "Registry transfer lock" },
    serverDeleteProhibited: { zh: "注册局禁止删除", en: "Registry delete lock" },
    serverUpdateProhibited: { zh: "注册局禁止更新", en: "Registry update lock" },
    serverRenewProhibited: { zh: "注册局禁止续费", en: "Registry renew lock" },
    serverHold: { zh: "注册局暂停", en: "Server hold" },
    pendingDelete: { zh: "等待删除", en: "Pending delete" },
    pendingTransfer: { zh: "等待转移", en: "Pending transfer" },
    pendingUpdate: { zh: "等待更新", en: "Pending update" },
    pendingCreate: { zh: "等待创建", en: "Pending create" },
    pendingRenew: { zh: "等待续费", en: "Pending renew" },
    pendingRestore: { zh: "等待恢复", en: "Pending restore" },
    redemptionPeriod: { zh: "赎回期", en: "Redemption" },
    addPeriod: { zh: "新增宽限期", en: "Add grace" },
    autoRenewPeriod: { zh: "自动续费宽限期", en: "Auto-renew grace" },
    renewPeriod: { zh: "续费宽限期", en: "Renew grace" },
    transferPeriod: { zh: "转移宽限期", en: "Transfer grace" }
  };

  var state = {
    lang: "zh",
    loading: false,
    result: null,
    error: null,
    showJson: false,
    recent: [],
    batch: [],
    batchLoading: false,
    batchError: null,
    batchTrim: false
  };
  var ctrl = null;
  var toastTimer = 0;

  function $(id) { return document.getElementById(id); }

  function el(tag, attrs, kids) {
    var node = document.createElement(tag);
    if (attrs) {
      Object.keys(attrs).forEach(function (k) {
        var v = attrs[k];
        if (v == null || v === false) return;
        if (k === "class") node.className = v;
        else if (k === "text") node.textContent = v;
        else node.setAttribute(k, v);
      });
    }
    (kids || []).forEach(function (kid) {
      if (kid == null || kid === false) return;
      node.append(kid.nodeType ? kid : document.createTextNode(String(kid)));
    });
    return node;
  }

  function t(key) {
    var pack = I18N[state.lang] || I18N.zh;
    return Object.prototype.hasOwnProperty.call(pack, key) ? pack[key] : key;
  }

  function dash(v) {
    return v == null || v === "" ? "—" : String(v);
  }

  function fmtDate(iso) {
    if (!iso || iso.indexOf("0001-01-01") === 0) return "";
    var d = new Date(iso);
    if (isNaN(d.getTime())) return String(iso);
    var formatted = new Intl.DateTimeFormat(state.lang === "en" ? "en-GB" : "zh-CN", {
      year: "numeric",
      month: "short",
      day: "numeric",
      timeZone: "UTC"
    }).format(d);
    return formatted + " UTC";
  }

  function gloss(code) {
    var row = STATUS_GLOSS[code];
    if (!row) return "";
    return state.lang === "en" ? row.en : row.zh;
  }

  function statusTone(code) {
    var k = String(code || "").toLowerCase();
    if (k === "ok" || k === "active") return "ok";
    if (k.indexOf("hold") >= 0 || k.indexOf("pending") >= 0 || k.indexOf("redemption") >= 0) return "warn";
    return "";
  }

  function safeHttp(url) {
    try {
      var u = new URL(url);
      if (u.protocol === "http:" || u.protocol === "https:") return u.href;
    } catch (e) { /* not a URL */ }
    return "";
  }

  function linkOrText(url) {
    if (!url) return "";
    var href = safeHttp(url);
    if (!href) return url;
    return el("a", { href: href, target: "_blank", rel: "noopener noreferrer", text: url });
  }

  function mailLink(addr) {
    if (!addr) return "";
    if (!/^[^@\s]+@[^@\s]+$/.test(addr)) return addr;
    return el("a", { href: "mailto:" + encodeURIComponent(addr), text: addr });
  }

  function kvRow(label, value) {
    var dd = el("dd");
    if (value == null || value === "") dd.textContent = "—";
    else if (typeof value === "string") {
      dd.textContent = value;
      if (value.length > 22) dd.title = value;
    } else dd.append(value);
    return el("div", { class: "kv-row" }, [el("dt", { text: label }), dd]);
  }

  function regText(yes) { return yes ? t("reg.yes") : t("reg.no"); }

  function sourceText(meta) {
    if (!meta || !meta.source) return "";
    var label = t("source." + meta.source);
    return label === "source." + meta.source ? meta.source : label;
  }

  function errorText(err) {
    if (!err) return t("err.generic");
    if (err.code && I18N[state.lang]["err." + err.code]) return t("err." + err.code);
    return err.message || t("err.generic");
  }

  function readStore(key, fallback) {
    try {
      var raw = localStorage.getItem(key);
      if (!raw) return fallback;
      return JSON.parse(raw);
    } catch (e) {
      return fallback;
    }
  }

  function writeStore(key, value) {
    try { localStorage.setItem(key, JSON.stringify(value)); } catch (e) { /* private mode */ }
  }

  function toast(msg) {
    var node = $("toast");
    node.textContent = msg;
    node.classList.add("show");
    clearTimeout(toastTimer);
    toastTimer = setTimeout(function () { node.classList.remove("show"); }, 1600);
  }

  function copyText(text) {
    function ok() { toast(t("copied")); }
    function fail() { toast(t("copyfail")); }
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text).then(ok, function () { fallback(); });
      return;
    }
    fallback();
    function fallback() {
      var ta = document.createElement("textarea");
      ta.value = text;
      ta.setAttribute("readonly", "");
      ta.style.position = "fixed";
      ta.style.left = "-9999px";
      document.body.append(ta);
      ta.select();
      try {
        if (document.execCommand("copy")) ok();
        else fail();
      } catch (e) { fail(); }
      ta.remove();
    }
  }

  function summaryText(r, brief) {
    if (!r) return "";
    var reg = r.registrar || {};
    var dates = r.dates || {};
    var lines = [r.domain || r.query || ""];
    lines.push(t("sum.registered") + ": " + regText(!!r.isRegistered));
    lines.push(t("sum.registrar") + ": " + (reg.name || "—"));
    lines.push(t("sum.created") + ": " + (fmtDate(dates.created) || "—"));
    if (!brief) lines.push(t("sum.updated") + ": " + (fmtDate(dates.updated) || "—"));
    lines.push(t("sum.expires") + ": " + (fmtDate(dates.expires) || "—"));
    var ns = (r.nameservers || []).map(function (n) { return n.name; }).filter(Boolean);
    if (ns.length) lines.push(t("sum.ns") + ": " + ns.join(", "));
    if (!brief && r.status && r.status.length) lines.push(t("sum.statuses") + ": " + r.status.join(", "));
    var src = sourceText(r.meta);
    if (src) lines.push(t("sum.source") + ": " + src);
    return lines.join("\n");
  }

  function applyI18n() {
    document.documentElement.lang = state.lang === "en" ? "en" : "zh-CN";
    document.title = t("title");
    var desc = document.querySelector('meta[name="description"]');
    if (desc) desc.setAttribute("content", t("desc"));
    document.querySelectorAll("[data-i18n]").forEach(function (node) {
      node.textContent = t(node.getAttribute("data-i18n"));
    });
    document.querySelectorAll("[data-i18n-html]").forEach(function (node) {
      node.innerHTML = t(node.getAttribute("data-i18n-html"));
    });
    document.querySelectorAll("[data-i18n-placeholder]").forEach(function (node) {
      node.setAttribute("placeholder", t(node.getAttribute("data-i18n-placeholder")));
    });
    var langBtn = $("langToggle");
    langBtn.setAttribute("aria-label", t("lang.aria"));
    var jsonBtn = $("jsonToggle");
    jsonBtn.textContent = state.showJson ? t("json.hide") : t("record.json");
    jsonBtn.setAttribute("aria-pressed", state.showJson ? "true" : "false");
  }

  function paintStatus() {
    var node = $("status");
    node.classList.remove("is-busy", "is-err");
    if (state.loading) {
      node.classList.add("is-busy");
      node.textContent = t("status.loading");
      return;
    }
    if (state.error) {
      node.classList.add("is-err");
      node.textContent = errorText(state.error);
      return;
    }
    if (state.result) {
      node.textContent = t("status.done").replace("{domain}", state.result.domain || state.result.query || "");
      return;
    }
    node.textContent = t("status.idle");
  }

  function paintBadge() {
    var b = $("snapBadge");
    if (state.loading) {
      b.textContent = t("snap.badge.loading");
      b.className = "badge";
      return;
    }
    if (state.result && state.result.meta && state.result.meta.source) {
      b.textContent = String(state.result.meta.source).toUpperCase();
      b.className = "badge badge-lime";
      return;
    }
    b.textContent = t("snap.badge.wait");
    b.className = "badge badge-grey";
  }

  function paintBar() {
    var q = $("barQuery");
    var dot = $("barDot");
    if (state.loading) dot.className = "dot dot-violet";
    else if (state.result && state.result.domain) dot.className = "dot dot-lime";
    else dot.className = "dot dot-grey";
    q.textContent = state.result && state.result.domain ? state.result.domain : t("bar.wait");
  }

  function paintSnap() {
    var r = state.result;
    var host = $("snapBody");
    var reg = (r && r.registrar) || {};
    var dates = (r && r.dates) || {};
    var name = r && (r.domain || r.query) ? (r.domain || r.query) : "— — —";
    var sub = r ? regText(!!r.isRegistered) : t("snap.sub.wait");
    var plate = el("div", { class: "plate" }, [
      el("div", { class: "plate-id" }, [
        el("div", { class: "plate-name", text: name }),
        el("div", { class: r && !r.isRegistered ? "plate-sub is-warn" : "plate-sub", text: sub })
      ]),
      r ? el("span", { class: "pill pill-sky", text: (r.tld || "").toUpperCase() || "—" }) : null
    ]);
    var kv = el("dl", { class: "kv" }, [
      kvRow(t("snap.f.status"), r ? regText(!!r.isRegistered) : ""),
      kvRow(t("snap.f.registrar"), reg.name || ""),
      kvRow(t("snap.f.created"), fmtDate(dates.created)),
      kvRow(t("snap.f.expires"), fmtDate(dates.expires)),
      kvRow(t("snap.f.source"), r ? sourceText(r.meta) : ""),
      kvRow(t("snap.f.tld"), r && r.tld ? r.tld : "")
    ]);
    host.replaceChildren(plate, kv);
  }

  function addressText(a) {
    if (!a) return "";
    return [a.street, a.city, a.state, a.postalCode, a.country].filter(Boolean).join(", ");
  }

  function paintRecord() {
    var host = $("recordBody");
    var raw = $("rawJson");
    var r = state.result;
    if (!r) {
      host.replaceChildren(el("div", { class: "empty" }, [
        el("strong", { text: t("record.empty.title") }),
        el("p", { text: t("record.empty") })
      ]));
      raw.hidden = true;
      raw.textContent = "";
      return;
    }
    var reg = r.registrar || {};
    var dates = r.dates || {};
    var sub = regText(!!r.isRegistered);
    if (reg.name) sub += " · " + reg.name;
    var top = el("div", { class: "res-top" }, [
      el("div", { class: "res-id" }, [
        el("div", { class: "res-name", text: r.domain || r.query || "—" }),
        el("div", { class: r.isRegistered ? "res-sub" : "res-sub is-warn", text: sub }),
        r.domainUnicode && r.domainUnicode !== r.domain
          ? el("div", { class: "res-path", text: r.domainUnicode })
          : null
      ]),
      el("span", { class: "pill", text: (r.tld || "").toUpperCase() || "—" })
    ]);

    var rows = [
      kvRow(t("record.domain"), r.domain || ""),
      r.domainUnicode && r.domainUnicode !== r.domain ? kvRow(t("record.unicode"), r.domainUnicode) : null,
      r.id ? kvRow(t("record.id"), r.id) : null,
      kvRow(t("record.tld"), r.tld || ""),
      kvRow(t("record.registered"), regText(!!r.isRegistered)),
      kvRow(t("record.registrar"), reg.name || ""),
      reg.ianaId ? kvRow(t("record.iana"), reg.ianaId) : null,
      reg.url ? kvRow(t("record.url"), linkOrText(reg.url)) : null,
      reg.whoisServer ? kvRow(t("record.whois"), reg.whoisServer) : null,
      reg.reseller ? kvRow(t("record.reseller"), reg.reseller) : null,
      reg.abuseEmail ? kvRow(t("record.abuseEmail"), mailLink(reg.abuseEmail)) : null,
      reg.abusePhone ? kvRow(t("record.abusePhone"), reg.abusePhone) : null,
      kvRow(t("record.created"), fmtDate(dates.created)),
      kvRow(t("record.updated"), fmtDate(dates.updated)),
      kvRow(t("record.expires"), fmtDate(dates.expires))
    ];
    var kv = el("dl", { class: "kv kv-2" }, rows);

    var codes = r.status || [];
    var altKids = [
      el("div", { class: "alt-title" }, [
        document.createTextNode(t("record.statuses")),
        el("span", { text: t("record.statuses.hint") })
      ])
    ];
    if (!codes.length) {
      altKids.push(el("p", { class: "fine", text: t("record.noStatus") }));
    } else {
      altKids.push(el("div", { class: "signals" }, codes.map(function (code) {
        var label = gloss(code) || code;
        return el("div", { class: "signal " + statusTone(code) }, [
          el("i"),
          el("span", { text: label }),
          gloss(code) ? el("span", { class: "code", text: code }) : null
        ]);
      })));
    }
    var alt = el("div", { class: "alt" }, altKids);

    var metaBits = [];
    if (r.meta) {
      var src = sourceText(r.meta);
      if (src) metaBits.push(src);
      if (r.meta.server) metaBits.push(r.meta.server);
      if (r.meta.cached) metaBits.push(t("cached"));
      var fetched = fmtDate(r.meta.fetchedAt);
      if (fetched) metaBits.push(fetched);
    }
    var meta = metaBits.length ? el("p", { class: "meta-line", text: metaBits.join(" · ") }) : null;
    host.replaceChildren(top, kv, alt, meta);

    if (state.showJson) {
      raw.hidden = false;
      raw.textContent = JSON.stringify(r, null, 2);
    } else {
      raw.hidden = true;
    }
  }

  function paintNS() {
    var host = $("nsBody");
    var r = state.result;
    if (!r) {
      host.replaceChildren(
        el("div", { class: "score" }, [
          el("div", { class: "score-num is-wait", text: "—" }),
          el("div", { class: "score-side" }, [
            el("div", { class: "score-label", text: t("dnssec.wait") }),
            el("p", { text: t("ns.empty") })
          ])
        ])
      );
      return;
    }
    var dns = r.dnssec || {};
    var signed = !!dns.signed;
    var nss = r.nameservers || [];
    var ds = dns.dsData || [];
    var score = el("div", { class: "score" }, [
      el("div", { class: signed ? "score-num" : "score-num is-off", text: signed ? (state.lang === "en" ? "ON" : "签") : (state.lang === "en" ? "OFF" : "未") }),
      el("div", { class: "score-side" }, [
        el("div", { class: "score-label", text: signed ? t("dnssec.on") : t("dnssec.off") }),
        el("p", { text: signed ? t("dnssec.detail.on") : t("dnssec.detail.off") })
      ])
    ]);
    var facts = el("dl", { class: "kv" }, [
      kvRow(t("dnssec.ns"), String(nss.length)),
      kvRow(t("dnssec.ds"), String(ds.length))
    ]);
    var list;
    if (!nss.length) {
      list = el("p", { class: "fine", text: t("ns.none") });
    } else {
      list = el("div", { class: "ns-list" }, nss.map(function (ns, i) {
        var ips = (ns.ipv4 || []).concat(ns.ipv6 || []);
        return el("div", { class: "ns" }, [
          el("span", { class: "ns-i", text: String(i + 1).padStart(2, "0") }),
          el("div", {}, [
            el("div", { class: "ns-name", text: ns.name || "—" }),
            ips.length ? el("div", { class: "ns-ip", text: ips.join("  ") }) : null
          ])
        ]);
      }));
    }
    var dsBox = null;
    if (ds.length) {
      dsBox = el("div", { class: "ns-list" }, ds.map(function (item) {
        var digest = item.digest || "";
        var short = digest.length > 24 ? digest.slice(0, 24) + "…" : digest;
        return el("div", { class: "ds", title: digest, text: "DS " + (item.keyTag || "") + " · alg " + (item.algorithm || "") + " · " + short });
      }));
    }
    host.replaceChildren(score, facts, list, dsBox);
  }

  function contactCard(role, c) {
    c = c || {};
    var addr = addressText(c.address);
    var fields = [];
    if (c.name) fields.push(kvRow(t("c.name"), c.name));
    if (c.organization) fields.push(kvRow(t("c.org"), c.organization));
    if (c.email) fields.push(kvRow(t("c.email"), mailLink(c.email)));
    if (c.phone) fields.push(kvRow(t("c.phone"), c.phone));
    if (addr) fields.push(kvRow(t("c.addr"), addr));
    var pill = null;
    if (!fields.length && c.redacted) pill = el("span", { class: "pill pill-pink", text: t("contacts.redacted") });
    else if (!fields.length) pill = el("span", { class: "pill pill-wait", text: t("contacts.none") });
    var body = fields.length
      ? el("dl", { class: "kv" }, fields)
      : el("p", { class: "fine", text: c.redacted ? t("contacts.redacted") : t("contacts.none") });
    return el("article", { class: "role" }, [
      el("div", { class: "role-h" }, [
        el("span", { text: t("role." + role) }),
        pill
      ]),
      body
    ]);
  }

  function paintContacts() {
    var host = $("contactsBody");
    var r = state.result;
    if (!r) {
      host.replaceChildren(el("div", { class: "empty" }, [
        el("strong", { text: t("contacts.title") }),
        el("p", { text: t("contacts.empty") })
      ]));
      return;
    }
    var c = r.contacts || {};
    host.replaceChildren(el("div", { class: "roles" }, [
      contactCard("registrant", c.registrant),
      contactCard("admin", c.admin),
      contactCard("tech", c.tech),
      contactCard("billing", c.billing)
    ]));
  }

  function domainTokens(raw) {
    var seen = Object.create(null);
    var out = [];
    String(raw || "").split(/[\s,;]+/).forEach(function (part) {
      var s = part.trim();
      if (!s) return;
      var k = s.toLowerCase();
      if (seen[k]) return;
      seen[k] = true;
      out.push(s);
    });
    return out;
  }

  function paintBatchCount() {
    var n = domainTokens($("batchInput").value).length;
    var node = $("batchCount");
    node.textContent = n + " / " + MAX_BATCH;
    node.classList.toggle("is-over", n > MAX_BATCH);
  }

  function paintBatch() {
    paintBatchCount();
    var run = $("batchRun");
    run.disabled = state.batchLoading;
    run.textContent = state.batchLoading ? t("batch.running") : t("batch.run");
    $("batchCsv").disabled = !state.batch.length;
    var status = $("batchStatus");
    status.classList.remove("is-busy", "is-err");
    if (state.batchLoading) {
      status.classList.add("is-busy");
      status.textContent = t("status.loading");
    } else if (state.batchError) {
      status.classList.add("is-err");
      status.textContent = state.batchError === "EMPTY" ? t("batch.need") : errorText(state.batchError);
    } else if (state.batchTrim) {
      status.textContent = t("batch.trim");
    } else {
      status.textContent = "";
    }

    var body = $("batchRows");
    if (!state.batch.length) {
      body.replaceChildren(el("tr", {}, [
        el("td", { class: "hint", colspan: "5", text: t("batch.empty") })
      ]));
      return;
    }
    var rows = state.batch.map(function (row) {
      if (row && row.error) {
        var q = row.error.query || "—";
        return el("tr", {}, [
          el("td", { class: "m", text: q }),
          el("td", {}, [el("span", { class: "pill pill-err", text: errorText(row.error) })]),
          el("td", { text: "—" }),
          el("td", { class: "m", text: "—" }),
          el("td", { class: "m", text: "—" })
        ]);
      }
      var reg = (row && row.registrar) || {};
      var dates = (row && row.dates) || {};
      var pill = row && row.isRegistered ? "pill pill-lime" : "pill pill-wait";
      return el("tr", {}, [
        el("td", { class: "m", text: (row && (row.domain || row.query)) || "—" }),
        el("td", {}, [el("span", { class: pill, text: row ? regText(!!row.isRegistered) : "—" })]),
        el("td", { text: reg.name || "—" }),
        el("td", { class: "m", text: fmtDate(dates.expires) || "—" }),
        el("td", { class: "m", text: sourceText(row && row.meta) || "—" })
      ]);
    });
    body.replaceChildren();
    rows.forEach(function (row) { body.append(row); });
  }

  function renderCurl() {
    var origin = location.origin;
    var root = $("curlOut");
    var lines = [
      { kind: "c", text: t("curl.c1") },
      { kind: "cmd", text: "curl " + origin + "/v1/whois/example.com" },
      { kind: "blank" },
      { kind: "c", text: t("curl.c2") },
      { kind: "cmd", text: "curl '" + origin + "/v1/whois/example.com?raw=true'" },
      { kind: "blank" },
      { kind: "c", text: t("curl.c3") },
      { kind: "cmd", text: "curl '" + origin + "/multi?domains=example.com,github.com'" }
    ];
    root.replaceChildren();
    lines.forEach(function (line) {
      if (line.kind === "blank") {
        root.append(document.createTextNode("\n"));
        return;
      }
      if (line.kind === "c") {
        root.append(el("span", { class: "t-c", text: "# " + line.text }), document.createTextNode("\n"));
        return;
      }
      root.append(
        el("span", { class: "t-p", text: "$ " }),
        el("span", { text: line.text }),
        document.createTextNode("\n")
      );
    });
  }

  function renderRecent() {
    var row = $("histRow");
    var host = $("hist");
    host.replaceChildren();
    if (!state.recent.length) {
      row.hidden = true;
      return;
    }
    row.hidden = false;
    state.recent.forEach(function (d) {
      host.append(el("button", { type: "button", class: "chip", "data-domain": d, text: d }));
    });
  }

  function remember(domain) {
    if (!domain) return;
    var key = domain.toLowerCase();
    state.recent = [domain].concat(state.recent.filter(function (d) { return d.toLowerCase() !== key; })).slice(0, 8);
    writeStore(RECENT_KEY, state.recent);
    renderRecent();
  }

  function paint() {
    paintStatus();
    paintBadge();
    paintBar();
    paintSnap();
    paintRecord();
    paintNS();
    paintContacts();
    $("searchBtn").disabled = state.loading;
    $("searchBtn").textContent = state.loading ? t("btn.searching") : t("btn.search");
    $("q").setAttribute("aria-busy", state.loading ? "true" : "false");
    $("snapCopy").disabled = !state.result;
    $("recordCopy").disabled = !state.result;
  }

  function fetchJSON(url, signal) {
    return fetch(url, { headers: { Accept: "application/json" }, signal: signal }).then(function (res) {
      return res.text().then(function (text) {
        var data = null;
        if (text) {
          try { data = JSON.parse(text); } catch (e) { data = null; }
        }
        if (!res.ok) {
          var err = data && data.error ? data.error : {};
          var wrapped = new Error(err.message || "");
          wrapped.code = err.code || "";
          wrapped.detail = err.detail || "";
          throw wrapped;
        }
        if (!data) {
          var bad = new Error("");
          bad.code = "NETWORK";
          throw bad;
        }
        return data;
      });
    }, function (err) {
      if (err && err.name === "AbortError") throw err;
      var net = new Error("");
      net.code = "NETWORK";
      throw net;
    });
  }

  function lookup(raw) {
    var q = String(raw || "").trim();
    $("q").value = q;
    if (!q) {
      if (ctrl) { ctrl.abort(); ctrl = null; }
      state.error = { code: "EMPTY" };
      state.loading = false;
      paint();
      return;
    }
    var prev = ctrl;
    ctrl = new AbortController();
    if (prev) prev.abort();
    var mine = ctrl;
    state.loading = true;
    state.error = null;
    paint();
    fetchJSON("/v1/whois/" + encodeURIComponent(q), mine.signal).then(function (data) {
      if (mine.signal.aborted) return;
      state.result = data;
      state.error = null;
      state.showJson = false;
      $("jsonToggle").textContent = t("record.json");
      $("jsonToggle").setAttribute("aria-pressed", "false");
      remember(data.domain || q);
    }, function (err) {
      if (err && err.name === "AbortError") return;
      if (mine.signal.aborted) return;
      state.error = err;
    }).then(function () {
      if (ctrl !== mine) return;
      state.loading = false;
      ctrl = null;
      paint();
      if (state.result && !state.error && window.matchMedia("(max-width: 980px)").matches) {
        var snap = $("snap");
        if (snap && snap.scrollIntoView) snap.scrollIntoView({ behavior: "smooth", block: "nearest" });
      }
    });
  }

  function runBatch() {
    var tokens = domainTokens($("batchInput").value);
    if (!tokens.length) {
      state.batchError = "EMPTY";
      state.batchTrim = false;
      paintBatch();
      return;
    }
    var list = tokens.slice(0, MAX_BATCH);
    state.batchTrim = tokens.length > MAX_BATCH;
    state.batchLoading = true;
    state.batchError = null;
    paintBatch();
    var url = "/multi?domains=" + list.map(encodeURIComponent).join(",");
    fetchJSON(url).then(function (data) {
      state.batch = data.results || [];
    }, function (err) {
      state.batchError = err;
      state.batch = [];
    }).then(function () {
      state.batchLoading = false;
      paintBatch();
    });
  }

  function csvCell(v) {
    var s = v == null ? "" : String(v);
    if (/[",\n]/.test(s)) return '"' + s.split('"').join('""') + '"';
    return s;
  }

  function exportCSV() {
    if (!state.batch.length) return;
    var lines = [["domain", "registered", "registrar", "created", "expires", "source", "error"].join(",")];
    state.batch.forEach(function (row) {
      if (row && row.error) {
        lines.push([
          csvCell(row.error.query), "", "", "", "", "", csvCell(row.error.code || row.error.message)
        ].join(","));
        return;
      }
      var reg = (row && row.registrar) || {};
      var dates = (row && row.dates) || {};
      lines.push([
        csvCell(row && row.domain),
        row && row.isRegistered ? "yes" : "no",
        csvCell(reg.name),
        csvCell(dates.created),
        csvCell(dates.expires),
        csvCell(row && row.meta && row.meta.source),
        ""
      ].join(","));
    });
    var blob = new Blob(["\uFEFF" + lines.join("\n")], { type: "text/csv;charset=utf-8" });
    var a = document.createElement("a");
    a.href = URL.createObjectURL(blob);
    a.download = "whois.csv";
    document.body.append(a);
    a.click();
    a.remove();
    setTimeout(function () { URL.revokeObjectURL(a.href); }, 1000);
  }

  function setLang(next) {
    state.lang = next === "en" ? "en" : "zh";
    writeStore(LANG_KEY, state.lang);
    applyI18n();
    renderCurl();
    renderRecent();
    paint();
    paintBatch();
  }

  function bind() {
    $("lookupForm").addEventListener("submit", function (e) {
      e.preventDefault();
      lookup($("q").value);
    });
    $("clearBtn").addEventListener("click", function () {
      $("q").value = "";
      $("q").focus();
    });
    document.body.addEventListener("click", function (e) {
      var chip = e.target.closest("[data-domain]");
      if (!chip) return;
      lookup(chip.getAttribute("data-domain"));
    });
    $("langToggle").addEventListener("click", function () {
      setLang(state.lang === "en" ? "zh" : "en");
    });
    $("barMe").addEventListener("click", function () {
      var target = state.result ? $("record") : $("lookup");
      if (target && target.scrollIntoView) target.scrollIntoView({ behavior: "smooth", block: "start" });
    });
    $("snapCopy").addEventListener("click", function () {
      if (state.result) copyText(summaryText(state.result, true));
    });
    $("recordCopy").addEventListener("click", function () {
      if (state.result) copyText(summaryText(state.result, false));
    });
    $("jsonToggle").addEventListener("click", function () {
      if (!state.result) { toast(t("json.need")); return; }
      state.showJson = !state.showJson;
      $("jsonToggle").textContent = state.showJson ? t("json.hide") : t("record.json");
      $("jsonToggle").setAttribute("aria-pressed", state.showJson ? "true" : "false");
      paintRecord();
    });
    $("clearRecent").addEventListener("click", function () {
      state.recent = [];
      writeStore(RECENT_KEY, []);
      renderRecent();
    });
    $("batchInput").addEventListener("input", paintBatchCount);
    $("batchRun").addEventListener("click", runBatch);
    $("batchClear").addEventListener("click", function () {
      $("batchInput").value = "";
      state.batch = [];
      state.batchError = null;
      state.batchTrim = false;
      paintBatch();
    });
    $("batchCsv").addEventListener("click", exportCSV);
  }

  var storedLang = readStore(LANG_KEY, null);
  if (storedLang === "en" || storedLang === "zh") state.lang = storedLang;
  var storedRecent = readStore(RECENT_KEY, []);
  if (Array.isArray(storedRecent)) {
    state.recent = storedRecent.filter(function (d) { return typeof d === "string" && d; }).slice(0, 8);
  }

  applyI18n();
  renderCurl();
  renderRecent();
  paint();
  paintBatch();
  bind();

  var initial = "";
  try { initial = new URLSearchParams(location.search).get("q") || ""; } catch (e) { initial = ""; }
  if (initial.trim()) lookup(initial.trim());
})();
