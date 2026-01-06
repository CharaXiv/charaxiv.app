// Component scripts - bundled for caching

// NumberInput adjustment (optimistic UI)
function adjustNumberInput(btn, delta) {
  const wrapper = btn.parentNode;
  const input = wrapper.querySelector('input');
  const value = parseInt(input.value) || 0;
  const min = parseInt(input.min);
  const max = parseInt(input.max);
  const newValue = Math.max(min, Math.min(max, value + delta));
  input.value = newValue;

  wrapper.querySelectorAll('button').forEach(b => {
    const d = parseInt(b.title);
    b.disabled = (d < 0 && newValue <= min) || (d > 0 && newValue >= max);
  });
}

// Header scroll behavior
(function() {
  const header = document.getElementById('header');
  const logo = document.getElementById('header-logo');
  if (!header) return;
  
  function updateHeader() {
    if (window.scrollY <= 0) {
      header.classList.add('bg-white');
      header.classList.remove('bg-white/10');
    } else {
      header.classList.remove('bg-white');
      header.classList.add('bg-white/10');
    }
  }
  window.addEventListener('scroll', updateHeader, { passive: true });
  updateHeader();

  // HTMX loading indicator
  if (logo) {
    let requestCount = 0;
    document.body.addEventListener('htmx:beforeRequest', () => {
      requestCount++;
      logo.classList.add('htmx-request');
    });
    document.body.addEventListener('htmx:afterRequest', () => {
      requestCount--;
      if (requestCount <= 0) {
        requestCount = 0;
        logo.classList.remove('htmx-request');
      }
    });
  }
})();

// EasyMDE markdown editor initialization
(function() {
  var baseToolbar = [
    'heading', 'quote', 'unordered-list', 'ordered-list',
    '|',
    'bold', 'italic', 'strikethrough',
    '|',
    'link'
  ];

  function initMarkdownEditors() {
    if (typeof EasyMDE === 'undefined') {
      setTimeout(initMarkdownEditors, 100);
      return;
    }
    
    document.querySelectorAll('textarea.markdown-editor:not([data-easymde-init])').forEach(function(textarea) {
      textarea.setAttribute('data-easymde-init', 'true');
      var isSecret = textarea.hasAttribute('data-secret');
      
      var toolbar = baseToolbar.slice();
      if (isSecret) {
        toolbar.push('|');
        toolbar.push({
          name: 'password',
          action: function(editor) {
            var password = prompt('秘匿メモのパスワードを設定してください');
            if (password !== null) {
              alert('パスワードを設定しました');
            }
          },
          className: 'password-button',
          title: 'パスワード設定'
        });
      }
      
      var mde = new EasyMDE({
        element: textarea,
        status: false,
        spellChecker: false,
        minHeight: '150px',
        unorderedListStyle: '-',
        promptURLs: true,
        promptTexts: {
          image: '画像のURL',
          link: 'リンク先URL'
        },
        toolbar: toolbar
      });

      // Sync EasyMDE changes to textarea and trigger input event for HTMX
      mde.codemirror.on('change', function() {
        textarea.value = mde.value();
        textarea.dispatchEvent(new Event('input', { bubbles: true }));
      });
    });
  }
  
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initMarkdownEditors);
  } else {
    initMarkdownEditors();
  }
  
  // Re-init on HTMX swaps
  document.body.addEventListener('htmx:afterSwap', initMarkdownEditors);
})();

// Markdown rendering for readonly mode
(function() {
  function renderMarkdown() {
    if (typeof marked === 'undefined') {
      setTimeout(renderMarkdown, 100);
      return;
    }
    
    document.querySelectorAll('[data-markdown]:not([data-rendered])').forEach(function(el) {
      el.setAttribute('data-rendered', 'true');
      var content = el.getAttribute('data-markdown');
      if (content) {
        el.innerHTML = marked.parse(content);
      }
    });
  }
  
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', renderMarkdown);
  } else {
    renderMarkdown();
  }
  
  document.body.addEventListener('htmx:afterSwap', renderMarkdown);
})();

// Secret memo toggle (blur/reveal)
(function() {
  function initSecretToggles() {
    document.querySelectorAll('.secret-toggle-btn:not([data-toggle-init])').forEach(function(btn) {
      btn.setAttribute('data-toggle-init', 'true');
      btn.addEventListener('click', function() {
        var container = btn.closest('.markdown-renderer');
        if (!container) return;
        var content = container.querySelector('.markdown-content');
        if (!content) return;
        
        // Toggle blur
        var isHidden = content.classList.contains('blur-sm');
        if (isHidden) {
          content.classList.remove('blur-sm', 'select-none', 'pointer-events-none');
          // Update icon to eye-slash (content is now visible)
          btn.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 640 512" fill="currentColor" class="w-4 h-4"><path d="M38.8 5.1C28.4-3.1 13.3-1.2 5.1 9.2S-1.2 34.7 9.2 42.9l592 464c10.4 8.2 25.5 6.3 33.7-4.1s6.3-25.5-4.1-33.7L525.6 386.7c39.6-40.6 66.4-86.1 79.9-118.4c3.3-7.9 3.3-16.7 0-24.6c-14.9-35.7-46.2-87.7-93-131.1C465.5 68.8 400.8 32 320 32c-68.2 0-125 26.3-169.3 60.8L38.8 5.1zM223.1 149.5C248.6 126.2 282.7 112 320 112c79.5 0 144 64.5 144 144c0 24.9-6.3 48.3-17.4 68.7L408 294.5c8.4-19.3 10.6-41.4 4.8-63.3c-11.1-41.5-47.8-69.4-88.6-71.1c-5.8-.2-9.2 6.1-7.4 11.7c2.1 6.4 3.3 13.2 3.3 20.3c0 10.2-2.4 19.8-6.6 28.3l-90.3-70.8zM373 389.9c-16.4 6.5-34.3 10.1-53 10.1c-79.5 0-144-64.5-144-144c0-6.9 .5-13.6 1.4-20.2L83.1 161.5C60.3 191.2 44 220.8 34.5 243.7c-3.3 7.9-3.3 16.7 0 24.6c14.9 35.7 46.2 87.7 93 131.1C174.5 443.2 239.2 480 320 480c47.8 0 89.9-12.9 126.2-32.5L373 389.9z"/></svg>';
          btn.title = '秘匿メモを隠す';
        } else {
          content.classList.add('blur-sm', 'select-none', 'pointer-events-none');
          // Update icon to eye (content is now hidden)
          btn.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 576 512" fill="currentColor" class="w-4 h-4"><path d="M288 32c-80.8 0-145.5 36.8-192.6 80.6C48.6 156 17.3 208 2.5 243.7c-3.3 7.9-3.3 16.7 0 24.6C17.3 304 48.6 356 95.4 399.4C142.5 443.2 207.2 480 288 480s145.5-36.8 192.6-80.6c46.8-43.5 78.1-95.4 93-131.1c3.3-7.9 3.3-16.7 0-24.6c-14.9-35.7-46.2-87.7-93-131.1C433.5 68.8 368.8 32 288 32zM144 256a144 144 0 1 1 288 0 144 144 0 1 1 -288 0zm144-64c0 35.3-28.7 64-64 64c-7.1 0-13.9-1.2-20.3-3.3c-5.5-1.8-11.9 1.6-11.7 7.4c.3 6.9 1.3 13.8 3.2 20.7c13.7 51.2 66.4 81.6 117.6 67.9s81.6-66.4 67.9-117.6c-11.1-41.5-47.8-69.4-88.6-71.1c-5.8-.2-9.2 6.1-7.4 11.7c2.1 6.4 3.3 13.2 3.3 20.3z"/></svg>';
          btn.title = '秘匿メモを表示';
        }
      });
    });
  }
  
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initSecretToggles);
  } else {
    initSecretToggles();
  }
  
  document.body.addEventListener('htmx:afterSwap', initSecretToggles);
})();


// Gallery modal open/close
function openGalleryModal() {
  var modal = document.getElementById('gallery-modal');
  if (modal) {
    modal.classList.remove('hidden');
    modal.classList.add('flex');
    document.body.style.overflow = 'hidden';
  }
}

function closeGalleryModal() {
  var modal = document.getElementById('gallery-modal');
  if (modal) {
    modal.classList.add('hidden');
    modal.classList.remove('flex');
    document.body.style.overflow = '';
  }
}

// Close modal on Escape key
document.addEventListener('keydown', function(e) {
  if (e.key === 'Escape') {
    closeGalleryModal();
  }
});
