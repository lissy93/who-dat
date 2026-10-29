/* Contact form. Same language toggle as the lookup page (localStorage whois-ui-lang). */
(function () {
  "use strict";

  var INBOX = "i@wi.fs.kg";
  var LANG_KEY = "whois-ui-lang";
  var EMAIL_RE = /^[^\s@,;:<>()]+@[^\s@,;:<>()]+\.[^\s@,;:<>()]+$/;

  var I18N = {
    zh: {
      title: "联系我们 · wi.fs.kg",
      desc: "给 wi.fs.kg 留一句话：域名查错了、接口想加字段、Bug 反馈都可以。不用注册，留言直接投到站长信箱，回信地址就是你自己填的邮箱。",
      "lang.aria": "切换到英文",
      "lang.btn": "EN",
      back: "回到查询台",
      eyebrow: "Contact · 直达信箱",
      "h1.pre": "有话直说，",
      "h1.mark": "一封信",
      "h1.post": "就到。",
      lede: "域名查错了、接口想加字段、页面哪里不好用，或者只是想吐槽一句——都写在这里。<b class=\"nb\">不用注册、不用验证码</b>，留言会直接投到站长的信箱，回信地址就是你自己填的那个邮箱。",
      "note.1": "<b>只送这一封信</b>：不会把你加进任何列表，也不会拿来发广告或做推送。",
      "note.2": "<b>带上了必要的上下文</b>：为了让问题能对上号，留言会附带提交时的页面地址与浏览器标识，方便复现。",
      "note.3": "<b>回信时间</b>：通常是看到就回，忙起来可能要一两天。技术问题请把复现步骤写清楚。",
      "tags.k": "快速起头",
      "tags.fine": "点一下会在留言框里带出提纲，按着填完直接发。",
      "tag.wrong": "数据不准",
      "tag.api": "接口建议",
      "tag.bug": "Bug 反馈",
      "tag.other": "随便聊聊",
      "field.name": "称呼",
      "field.opt": "选填",
      "field.email": "邮箱",
      "field.req": "必填",
      "field.msg": "留言",
      "field.human": "安全校验",
      "field.auto": "自动完成",
      "ph.name": "怎么称呼你",
      "ph.email": "you@example.com",
      "ph.msg": "想说什么都可以，越具体越好。",
      "ts.off": "这个部署还没打开自动校验",
      "hint.auto": "通常一秒内自动通过，不用点图片、不用输字符。只有可疑流量才会被要求手动确认一次。",
      "hint.blocked": "安全校验没加载出来，多半是广告拦截插件拦住了。关掉它再刷新；不想折腾就直接写信到 i@wi.fs.kg。",
      "hint.nokey": "这个部署还没打开安全校验。表单可以填，发不出去就直接写信到 i@wi.fs.kg。",
      "hint.nomail": "发信通道还没配好，提交会走不通。直接写信到 i@wi.fs.kg，我看到的时间一样。",
      send: "发送留言",
      sending: "发送中…",
      busy: "正在投递…",
      consent: "提交即表示同意我们仅为回复此留言而使用你填写的邮箱。内容不会公开展示。",
      "alt.1": "表单打不开或者提交不了？",
      "alt.2": "直接写信到 ",
      "alt.3": "，我看到的时间一样。",
      subject: "[wi.fs.kg] 来自联系页面",
      "done.title": "已经送到了",
      "done.p1": "谢谢你的留言。站长会直接回信到",
      "done.p2": "——记得查收，也顺手看一眼垃圾箱。",
      "done.again": "再写一封",
      "done.fallback": "你填的邮箱",
      "pn.lookup.t": "回查询台",
      "pn.lookup.d": "查域名，注册状态、注册商、到期日一次摊开。",
      "pn.docs.t": "接口文档",
      "pn.docs.d": "字段、错误码和调用示例都在这儿。",
      "pn.api.t": "终端接口",
      "pn.api.d": "curl 一行取 JSON，/multi 一次查多个。",
      "nav.contact": "联系",
      "nav.docs": "文档",
      "foot.blurb": "域名注册信息",
      "foot.privacy": "无广告 · 无追踪",
      "err.INVALID_EMAIL": "邮箱格式看起来不对",
      "err.MESSAGE_TOO_SHORT": "再多写几个字吧",
      "err.MESSAGE_TOO_LONG": "留言太长了",
      "err.BAD_JSON": "没读懂这条留言，刷新后再试",
      "err.REJECTED": "这条留言没能收下",
      "err.TURNSTILE_REQUIRED": "上面的安全校验还没通过，稍等一下再发送",
      "err.TURNSTILE_FAILED": "安全校验没通过。也可以直接写信到 i@wi.fs.kg",
      "err.NOT_CONFIGURED": "发信通道还没配好。可以直接写信到 i@wi.fs.kg",
      "err.RATE_LIMITED": "发得太勤了，稍后再试；也可以直接写信到 i@wi.fs.kg",
      "err.SEND_FAILED": "这会儿没送出去。可以直接写信到 i@wi.fs.kg",
      "err.NETWORK": "网络好像不太通，请稍后重试；也可以直接写信到 i@wi.fs.kg",
      tpl: {
        wrong: "查询的域名：\n期望的结果：\n实际显示：\n数据来源（RDAP 或 WHOIS）：\n",
        api: "想加的字段或能力：\n用在什么场景：\n示例域名：\n",
        bug: "出问题的页面：\n复现步骤：\n1. \n2. \n查询的域名：\n浏览器与系统：\n",
        other: "想说的：\n"
      }
    },
    en: {
      title: "Contact · wi.fs.kg",
      desc: "Write to wi.fs.kg: a wrong WHOIS record, an API idea, or a bug. No account. The reply goes to the address you type.",
      "lang.aria": "切换到中文",
      "lang.btn": "中文",
      back: "Back to lookup",
      eyebrow: "Contact · straight to the inbox",
      "h1.pre": "Say it plain. ",
      "h1.mark": "One letter",
      "h1.post": " gets there.",
      lede: "A record came back wrong, the API is missing a field, a page is fighting you, or you just want to grumble — write it here. <b class=\"nb\">No account, no captcha</b>. The note goes straight to the maintainer, and the reply goes to the address you type.",
      "note.1": "<b>Just this one letter</b>: you are not added to a list, and this is not used for ads or push.",
      "note.2": "<b>Enough context to match it up</b>: the page address and browser come along, so a bug can be reproduced.",
      "note.3": "<b>Reply time</b>: usually when it is seen, a day or two when things are busy. For anything technical, include the steps.",
      "tags.k": "Start from",
      "tags.fine": "A tap drops an outline into the box. Fill it in and send.",
      "tag.wrong": "Wrong data",
      "tag.api": "API idea",
      "tag.bug": "Bug",
      "tag.other": "Just saying hi",
      "field.name": "Name",
      "field.opt": "optional",
      "field.email": "Email",
      "field.req": "required",
      "field.msg": "Message",
      "field.human": "Security check",
      "field.auto": "automatic",
      "ph.name": "What should I call you",
      "ph.email": "you@example.com",
      "ph.msg": "Whatever it is, the more specific the better.",
      "ts.off": "Automatic check is not switched on here",
      "hint.auto": "Usually clears in a second. No pictures, no characters to type. Only suspicious traffic is asked to confirm.",
      "hint.blocked": "The security check did not load, usually because an ad blocker stopped it. Allow it and refresh, or just write to i@wi.fs.kg.",
      "hint.nokey": "This deploy has not switched on the security check. You can still fill the form; if it will not send, write to i@wi.fs.kg.",
      "hint.nomail": "Outgoing mail is not configured, so submit will not go through. Write to i@wi.fs.kg and I will see it just the same.",
      send: "Send note",
      sending: "Sending…",
      busy: "On its way…",
      consent: "Sending means this address is used only to reply. The note is not shown in public.",
      "alt.1": "Form will not open, or it will not send?",
      "alt.2": "Write to ",
      "alt.3": ". I will see it just the same.",
      subject: "[wi.fs.kg] from the contact page",
      "done.title": "It has been delivered",
      "done.p1": "Thanks for the note. The reply goes straight to",
      "done.p2": ". Check the inbox, and the spam folder too.",
      "done.again": "Write another",
      "done.fallback": "the address you typed",
      "pn.lookup.t": "Back to lookup",
      "pn.lookup.d": "Registration, registrar, and expiry, opened in one view.",
      "pn.docs.t": "API reference",
      "pn.docs.d": "Fields, error codes, and examples live here.",
      "pn.api.t": "Terminal API",
      "pn.api.d": "One curl returns JSON. /multi looks up several at once.",
      "nav.contact": "Contact",
      "nav.docs": "Docs",
      "foot.blurb": "Domain registration data",
      "foot.privacy": "No ads · no tracking",
      "err.INVALID_EMAIL": "That email looks wrong",
      "err.MESSAGE_TOO_SHORT": "Write a few more words",
      "err.MESSAGE_TOO_LONG": "That note is too long",
      "err.BAD_JSON": "Could not read that note. Refresh and try again",
      "err.REJECTED": "This note could not be accepted",
      "err.TURNSTILE_REQUIRED": "The security check above has not finished. Wait a moment, then send",
      "err.TURNSTILE_FAILED": "The security check did not pass. You can also write to i@wi.fs.kg",
      "err.NOT_CONFIGURED": "Outgoing mail is not configured. Write to i@wi.fs.kg instead",
      "err.RATE_LIMITED": "That is a lot of notes. Try again shortly, or write to i@wi.fs.kg",
      "err.SEND_FAILED": "It did not go out just now. Write to i@wi.fs.kg instead",
      "err.NETWORK": "The network looks unhappy. Try again, or write to i@wi.fs.kg",
      tpl: {
        wrong: "Domain looked up:\nExpected:\nWhat it showed:\nSource (RDAP or WHOIS):\n",
        api: "Field or capability:\nWhere it would be used:\nExample domain:\n",
        bug: "Page:\nSteps to reproduce:\n1. \n2. \nDomain:\nBrowser and OS:\n",
        other: "Note:\n"
      }
    }
  };

  var form = document.getElementById("cform");
  if (!form) return;

  var nameEl = document.getElementById("cName");
  var mailEl = document.getElementById("cEmail");
  var msgEl = document.getElementById("cMsg");
  var siteEl = document.getElementById("cSite");
  var btn = document.getElementById("cSend");
  var sendTxt = document.getElementById("cSendTxt");
  var stateEl = document.getElementById("cState");
  var countEl = document.getElementById("cCount");
  var doneEl = document.getElementById("cDone");
  var doneTo = document.getElementById("cDoneTo");
  var againEl = document.getElementById("cAgain");
  var humanHint = document.getElementById("cHumanHint");

  var state = {
    lang: "zh",
    siteKey: "",
    ready: false,
    cfgLoaded: false,
    widgetId: null,
    widgetLang: "",
    sending: false,
    tsBlocked: false,
    notice: null
  };

  function $(id) { return document.getElementById(id); }

  function t(key) {
    var pack = I18N[state.lang] || I18N.zh;
    return Object.prototype.hasOwnProperty.call(pack, key) ? pack[key] : key;
  }

  function readStore(key) {
    try {
      var raw = localStorage.getItem(key);
      if (!raw) return null;
      return JSON.parse(raw);
    } catch (e) {
      return null;
    }
  }

  function writeStore(key, value) {
    try { localStorage.setItem(key, JSON.stringify(value)); } catch (e) { /* private mode */ }
  }

  function setState(msg, kind, code) {
    state.notice = msg ? { kind: kind || "", code: code || "", text: msg } : null;
    paintNotice();
  }

  function paintNotice() {
    if (!stateEl) return;
    var n = state.notice;
    if (!n) {
      stateEl.textContent = "";
      stateEl.className = "cform-state";
      return;
    }
    var msg = n.text;
    if (n.kind === "busy") msg = t("busy");
    else if (n.code === "BLOCKED") msg = t("hint.blocked");
    else if (n.code && t("err." + n.code) !== "err." + n.code) msg = t("err." + n.code);
    stateEl.textContent = msg;
    stateEl.className = "cform-state" + (n.kind ? " " + n.kind : "");
  }

  function countUp() {
    if (!countEl || !msgEl) return;
    var n = msgEl.value.length;
    countEl.textContent = n + " / 5000";
    countEl.classList.toggle("full", n >= 4900);
  }

  function clearBad(el) {
    var f = el && el.parentNode;
    if (f && f.classList) f.classList.remove("bad");
    if (el) el.removeAttribute("aria-invalid");
  }

  function bad(el, code) {
    var f = el && el.parentNode;
    if (f && f.classList && f.classList.contains("field")) f.classList.add("bad");
    if (el) el.setAttribute("aria-invalid", "true");
    setState(t("err." + code), "err", code);
    if (el && el.focus) el.focus();
  }

  function paintHint() {
    if (!humanHint) return;
    humanHint.style.color = "";
    if (state.tsBlocked) {
      humanHint.textContent = t("hint.blocked");
      humanHint.style.color = "var(--coral)";
      return;
    }
    if (!state.cfgLoaded) {
      humanHint.textContent = t("hint.auto");
      return;
    }
    if (!state.siteKey) {
      humanHint.textContent = t("hint.nokey");
      return;
    }
    if (!state.ready) {
      humanHint.textContent = t("hint.nomail");
      return;
    }
    humanHint.textContent = t("hint.auto");
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
    if (langBtn) langBtn.setAttribute("aria-label", t("lang.aria"));
    var link = $("mailtoLink");
    if (link) link.href = "mailto:" + INBOX + "?subject=" + encodeURIComponent(t("subject"));
    if (sendTxt && !state.sending) sendTxt.textContent = t("send");
    paintHint();
    paintNotice();
    if (state.siteKey && state.widgetId != null && state.widgetLang !== state.lang) renderWidget();
  }

  function setLang(next) {
    state.lang = next === "en" ? "en" : "zh";
    writeStore(LANG_KEY, state.lang);
    applyI18n();
  }

  function renderWidget() {
    if (!state.siteKey || !window.turnstile || typeof window.turnstile.render !== "function") return false;
    var slot = $("tsWidget");
    if (!slot) return false;
    if (state.widgetId != null) {
      try { window.turnstile.remove(state.widgetId); } catch (e) { /* already gone */ }
      state.widgetId = null;
    }
    try {
      state.widgetId = window.turnstile.render(slot, {
        sitekey: state.siteKey,
        theme: "light",
        size: "flexible",
        language: state.lang === "en" ? "en" : "zh-cn"
      });
    } catch (e) {
      state.tsBlocked = true;
      paintHint();
      return false;
    }
    state.widgetLang = state.lang;
    state.tsBlocked = false;
    paintHint();
    return true;
  }

  function armBlockWatch() {
    var n = 0;
    var timer = setInterval(function () {
      if (state.widgetId != null) { clearInterval(timer); return; }
      if (window.turnstile && renderWidget()) { clearInterval(timer); return; }
      n += 1;
      if (n >= 40 && state.widgetId == null) {
        clearInterval(timer);
        state.tsBlocked = true;
        paintHint();
      }
    }, 100);
  }

  function applyConfig(cfg) {
    state.cfgLoaded = true;
    state.siteKey = (cfg && cfg.siteKey) || "";
    state.ready = !!(cfg && cfg.ready);
    var slot = $("tsSlot");
    var fallback = $("tsFallback");
    if (state.siteKey) {
      if (slot) slot.classList.remove("is-off");
      if (fallback) fallback.hidden = true;
      if (!renderWidget()) armBlockWatch();
    } else if (slot) {
      slot.classList.add("is-off");
      if (fallback) fallback.hidden = false;
    }
    paintHint();
  }

  function tsToken() {
    var hidden = form.querySelector('[name="cf-turnstile-response"]');
    if (hidden && hidden.value) return hidden.value;
    if (window.turnstile && state.widgetId != null && typeof window.turnstile.getResponse === "function") {
      try { return window.turnstile.getResponse(state.widgetId) || ""; } catch (e) { return ""; }
    }
    return "";
  }

  function tsReset() {
    if (!window.turnstile || state.widgetId == null || typeof window.turnstile.reset !== "function") return;
    try { window.turnstile.reset(state.widgetId); } catch (e) { /* widget already reset */ }
  }

  function errText(code, fallback) {
    if (code && t("err." + code) !== "err." + code) return t("err." + code);
    if (fallback && /[^\x00-\x7F]/.test(fallback) && state.lang === "zh") return fallback;
    if (fallback && state.lang === "en" && !/[^\x00-\x7F]/.test(fallback)) return fallback;
    return t("err.NETWORK");
  }

  function insertTpl(tpl) {
    if (!tpl) return;
    var v = msgEl.value;
    var s = msgEl.selectionStart;
    var e = msgEl.selectionEnd;
    if (typeof s !== "number") { s = e = v.length; }
    var pre = v.slice(0, s);
    var post = v.slice(e);
    var add = tpl;
    if (pre && !/\n$/.test(pre)) add = "\n" + add;
    msgEl.value = (pre + add + post).slice(0, 5000);
    var pos = Math.min((pre + add).length, msgEl.value.length);
    msgEl.focus();
    try { msgEl.setSelectionRange(pos, pos); } catch (err) { /* some browsers */ }
    countUp();
    clearBad(msgEl);
    if (stateEl && stateEl.classList.contains("err")) setState("");
  }

  msgEl.addEventListener("input", countUp);
  countUp();

  [nameEl, mailEl, msgEl].forEach(function (el) {
    if (!el) return;
    el.addEventListener("input", function () {
      clearBad(el);
      if (stateEl && stateEl.classList.contains("err")) setState("");
    });
  });

  Array.prototype.forEach.call(document.querySelectorAll(".ctag"), function (b) {
    b.addEventListener("click", function () {
      var pack = I18N[state.lang] || I18N.zh;
      insertTpl((pack.tpl && pack.tpl[b.getAttribute("data-k")]) || "");
    });
  });

  form.addEventListener("submit", function (e) {
    e.preventDefault();
    if (state.sending) return;

    var name = nameEl.value.trim();
    var email = mailEl.value.trim();
    var message = msgEl.value.trim();

    if (!EMAIL_RE.test(email)) return bad(mailEl, "INVALID_EMAIL");
    if (message.length < 5) return bad(msgEl, "MESSAGE_TOO_SHORT");
    if (message.length > 5000) return bad(msgEl, "MESSAGE_TOO_LONG");

    var token = tsToken();
    if (state.siteKey && state.ready && !token) {
      if (state.tsBlocked) setState(t("hint.blocked"), "err", "BLOCKED");
      else setState(t("err.TURNSTILE_REQUIRED"), "err", "TURNSTILE_REQUIRED");
      var slot = document.querySelector(".ts-slot");
      if (slot && slot.scrollIntoView) slot.scrollIntoView({ behavior: "smooth", block: "center" });
      return;
    }

    state.sending = true;
    btn.disabled = true;
    form.setAttribute("aria-busy", "true");
    sendTxt.textContent = t("sending");
    setState(t("busy"), "busy");

    fetch("/api/contact", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({
        name: name,
        email: email,
        message: message,
        website: siteEl ? siteEl.value : "",
        turnstile: token,
        page: location.pathname + location.search,
        lang: state.lang
      })
    }).then(function (r) {
      return r.json().catch(function () { return {}; }).then(function (d) {
        return { ok: r.ok, status: r.status, data: d || {} };
      });
    }).then(function (res) {
      if (!res.ok || !res.data.ok) {
        var code = res.data.code || "";
        var failure = new Error(errText(code, res.data.error));
        failure.code = code;
        throw failure;
      }
      if (doneTo) doneTo.textContent = email;
      form.reset();
      tsReset();
      countUp();
      setState("");
      form.classList.add("is-done");
      if (doneEl) doneEl.hidden = false;
    }).catch(function (err) {
      var m = (err && err.message) || "";
      var code = (err && err.code) || "";
      if (!m || m === "Failed to fetch" || m === "Load failed" || m === "NetworkError when attempting to fetch resource.") {
        m = t("err.NETWORK");
        code = "NETWORK";
      }
      setState(m, "err", code);
    }).then(function () {
      state.sending = false;
      btn.disabled = false;
      form.removeAttribute("aria-busy");
      sendTxt.textContent = t("send");
    });
  });

  if (againEl) {
    againEl.addEventListener("click", function () {
      form.classList.remove("is-done");
      if (doneEl) doneEl.hidden = true;
      if (doneTo) doneTo.textContent = t("done.fallback");
      tsReset();
      setState("");
      nameEl.focus();
    });
  }

  var stored = readStore(LANG_KEY);
  if (stored === "en" || stored === "zh") state.lang = stored;
  applyI18n();

  $("langToggle").addEventListener("click", function () {
    setLang(state.lang === "en" ? "zh" : "en");
  });

  fetch("/api/contact/config").then(function (r) {
    return r.json();
  }).then(applyConfig).catch(function () {
    state.cfgLoaded = true;
    paintHint();
  });
})();
