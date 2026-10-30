// ==============================================================================
// PolyTuring Master Documentation Portal — Client Logic & Interactive Engines
// Modern Day UX: Liquid Glass, 4-Theme Engine, Haptic Audio, Sidebar Resizer &
// Rail Collapse, Live Search Filter, Theater Mode Workbench, 1:1 ScrollSpy & TOC
// ==============================================================================

(function () {
  'use strict';

  // Global UI State
  const STATE = {
    theme: localStorage.getItem('polyturing_theme') || 'liquid',
    audioEnabled: localStorage.getItem('polyturing_audio_enabled') !== 'false',
    isSidebarCollapsed: localStorage.getItem('polyturing_sidebar_collapsed') === 'true',
    sidebarWidth: parseInt(localStorage.getItem('polyturing_sidebar_width'), 10) || 320,
    isFocusModeActive: false,
    isBatchExpanding: false,
    isNavClicking: false
  };

  // ---------------------------------------------------------------------------
  // 1. HAPTIC WEB AUDIO FEEDBACK ENGINE (Pure Web Audio API Synthesis)
  // ---------------------------------------------------------------------------
  let audioCtx = null;
  let lastAudioTimestamp = 0;

  const HapticAudio = {
    canPlay() {
      return STATE.audioEnabled && !STATE.isBatchExpanding && !!(window.AudioContext || window.webkitAudioContext);
    },

    getContext() {
      if (!audioCtx) {
        audioCtx = new (window.AudioContext || window.webkitAudioContext)();
      }
      if (audioCtx.state === 'suspended') {
        audioCtx.resume();
      }
      return audioCtx;
    },

    throttle(intervalMs = 18) {
      const now = performance.now();
      if (now - lastAudioTimestamp < intervalMs) return true;
      lastAudioTimestamp = now;
      return false;
    },

    // 1. Ascending Micro-Chirp for EXPAND (Chapters, Trees)
    expand() {
      if (!this.canPlay() || this.throttle(18)) return;
      try {
        const ctx = this.getContext();
        const now = ctx.currentTime;
        const osc = ctx.createOscillator();
        const gain = ctx.createGain();
        osc.type = 'sine';
        osc.frequency.setValueAtTime(540, now);
        osc.frequency.exponentialRampToValueAtTime(880, now + 0.038);
        gain.gain.setValueAtTime(0.045, now);
        gain.gain.exponentialRampToValueAtTime(0.0005, now + 0.038);
        osc.connect(gain);
        gain.connect(ctx.destination);
        osc.start(now);
        osc.stop(now + 0.04);
      } catch (e) {}
    },

    // 2. Descending Micro-Tuck for COLLAPSE (Chapters, Trees)
    collapse() {
      if (!this.canPlay() || this.throttle(18)) return;
      try {
        const ctx = this.getContext();
        const now = ctx.currentTime;
        const osc = ctx.createOscillator();
        const gain = ctx.createGain();
        osc.type = 'sine';
        osc.frequency.setValueAtTime(780, now);
        osc.frequency.exponentialRampToValueAtTime(420, now + 0.038);
        gain.gain.setValueAtTime(0.04, now);
        gain.gain.exponentialRampToValueAtTime(0.0005, now + 0.038);
        osc.connect(gain);
        gain.connect(ctx.destination);
        osc.start(now);
        osc.stop(now + 0.04);
      } catch (e) {}
    },

    // 3. Sliding Panel Open Glide (Sidebar / Drawer)
    sidebarOpen() {
      if (!this.canPlay()) return;
      try {
        const ctx = this.getContext();
        const now = ctx.currentTime;
        const osc = ctx.createOscillator();
        const gain = ctx.createGain();
        osc.type = 'triangle';
        osc.frequency.setValueAtTime(440, now);
        osc.frequency.exponentialRampToValueAtTime(760, now + 0.055);
        gain.gain.setValueAtTime(0.05, now);
        gain.gain.exponentialRampToValueAtTime(0.0005, now + 0.055);
        osc.connect(gain);
        gain.connect(ctx.destination);
        osc.start(now);
        osc.stop(now + 0.06);
      } catch (e) {}
    },

    // 4. Solid Mechanical Latch Click (Sidebar / Drawer Dismiss)
    sidebarClose() {
      if (!this.canPlay()) return;
      try {
        const ctx = this.getContext();
        const now = ctx.currentTime;
        const osc = ctx.createOscillator();
        const gain = ctx.createGain();
        osc.type = 'sine';
        osc.frequency.setValueAtTime(680, now);
        osc.frequency.exponentialRampToValueAtTime(320, now + 0.045);
        gain.gain.setValueAtTime(0.045, now);
        gain.gain.exponentialRampToValueAtTime(0.0005, now + 0.045);
        osc.connect(gain);
        gain.connect(ctx.destination);
        osc.start(now);
        osc.stop(now + 0.05);
      } catch (e) {}
    },

    // 5. Tactile Segmented / Chip Tap (Buttons, Mode Switchers, Links)
    tap() {
      if (!this.canPlay() || this.throttle(18)) return;
      try {
        const ctx = this.getContext();
        const now = ctx.currentTime;
        const osc = ctx.createOscillator();
        const gain = ctx.createGain();
        osc.type = 'triangle';
        osc.frequency.setValueAtTime(920, now);
        osc.frequency.exponentialRampToValueAtTime(560, now + 0.022);
        gain.gain.setValueAtTime(0.038, now);
        gain.gain.exponentialRampToValueAtTime(0.0005, now + 0.022);
        osc.connect(gain);
        gain.connect(ctx.destination);
        osc.start(now);
        osc.stop(now + 0.025);
      } catch (e) {}
    },

    // 6. Dual-Tone Celebration Chime (Code Copied)
    copySuccess() {
      if (!this.canPlay()) return;
      try {
        const ctx = this.getContext();
        const now = ctx.currentTime;
        const osc1 = ctx.createOscillator();
        const gain1 = ctx.createGain();
        osc1.type = 'sine';
        osc1.frequency.setValueAtTime(1046, now); // C6
        gain1.gain.setValueAtTime(0.04, now);
        gain1.gain.exponentialRampToValueAtTime(0.0005, now + 0.05);
        osc1.connect(gain1);
        gain1.connect(ctx.destination);
        osc1.start(now);
        osc1.stop(now + 0.05);

        const osc2 = ctx.createOscillator();
        const gain2 = ctx.createGain();
        osc2.type = 'sine';
        osc2.frequency.setValueAtTime(1318, now + 0.035); // E6
        gain2.gain.setValueAtTime(0.045, now + 0.035);
        gain2.gain.exponentialRampToValueAtTime(0.0005, now + 0.09);
        osc2.connect(gain2);
        gain2.connect(ctx.destination);
        osc2.start(now + 0.035);
        osc2.stop(now + 0.09);
      } catch (e) {}
    },

    // 7. Holographic Search Trigger
    searchOpen() {
      if (!this.canPlay()) return;
      try {
        const ctx = this.getContext();
        const now = ctx.currentTime;
        const osc = ctx.createOscillator();
        const gain = ctx.createGain();
        osc.type = 'sine';
        osc.frequency.setValueAtTime(1180, now);
        osc.frequency.exponentialRampToValueAtTime(1480, now + 0.04);
        gain.gain.setValueAtTime(0.04, now);
        gain.gain.exponentialRampToValueAtTime(0.0005, now + 0.045);
        osc.connect(gain);
        gain.connect(ctx.destination);
        osc.start(now);
        osc.stop(now + 0.05);
      } catch (e) {}
    },

    // 8. Cosmic Theme Cycle Sweep
    themeChange() {
      if (!this.canPlay()) return;
      try {
        const ctx = this.getContext();
        const now = ctx.currentTime;
        const osc = ctx.createOscillator();
        const gain = ctx.createGain();
        osc.type = 'triangle';
        osc.frequency.setValueAtTime(600, now);
        osc.frequency.exponentialRampToValueAtTime(1200, now + 0.06);
        gain.gain.setValueAtTime(0.045, now);
        gain.gain.exponentialRampToValueAtTime(0.0005, now + 0.06);
        osc.connect(gain);
        gain.connect(ctx.destination);
        osc.start(now);
        osc.stop(now + 0.065);
      } catch (e) {}
    },

    // 9. Visual Theater Workbench Launch
    theaterOpen() {
      if (!this.canPlay()) return;
      try {
        const ctx = this.getContext();
        const now = ctx.currentTime;
        const osc = ctx.createOscillator();
        const gain = ctx.createGain();
        osc.type = 'sine';
        osc.frequency.setValueAtTime(440, now);
        osc.frequency.exponentialRampToValueAtTime(880, now + 0.07);
        gain.gain.setValueAtTime(0.05, now);
        gain.gain.exponentialRampToValueAtTime(0.0005, now + 0.07);
        osc.connect(gain);
        gain.connect(ctx.destination);
        osc.start(now);
        osc.stop(now + 0.075);
      } catch (e) {}
    }
  };

  // ---------------------------------------------------------------------------
  // INITIALIZATION ON DOM READY
  // ---------------------------------------------------------------------------
  document.addEventListener('DOMContentLoaded', () => {
    initThemeEngine();
    initAudioToggle();
    initReadingProgressBar();
    initMathRendering();
    initSidebarResizer();
    initSidebarCollapseRail();
    initSidebarFilters();
    initMobileDrawer();
    initTreeNavigation();
    initCollapsibleChapters();
    initCanvasZoomAndTheater();
    initScrollSpy();
    initCopyButtons();
    initSearchModal();
    initTokenSimulator();
    initApiPlayground();
    initMermaid();
  });

  // ---------------------------------------------------------------------------
  // 2. MULTI-THEME ENGINE
  // ---------------------------------------------------------------------------
  function initThemeEngine() {
    const savedTheme = localStorage.getItem('polyturing_theme') || 'liquid';
    applyTheme(savedTheme, false);

    document.querySelectorAll('.theme-pill-btn').forEach(btn => {
      btn.addEventListener('click', () => {
        const targetTheme = btn.getAttribute('data-theme-set');
        if (targetTheme) {
          applyTheme(targetTheme, true);
        }
      });
    });
  }

  function applyTheme(themeName, playAudio = true) {
    STATE.theme = themeName;
    document.documentElement.setAttribute('data-theme', themeName);
    localStorage.setItem('polyturing_theme', themeName);

    document.querySelectorAll('.theme-pill-btn').forEach(btn => {
      if (btn.getAttribute('data-theme-set') === themeName) {
        btn.classList.add('active');
      } else {
        btn.classList.remove('active');
      }
    });

    if (playAudio) {
      HapticAudio.themeChange();
    }
  }

  // ---------------------------------------------------------------------------
  // 3. AUDIO FEEDBACK HUD TOGGLE
  // ---------------------------------------------------------------------------
  function initAudioToggle() {
    const toggleBtn = document.getElementById('audio-toggle-btn');
    if (!toggleBtn) return;

    function updateBtn() {
      toggleBtn.innerHTML = STATE.audioEnabled ? '🔊' : '🔇';
      toggleBtn.setAttribute('title', STATE.audioEnabled ? 'Haptic Audio: ON (Click to mute)' : 'Haptic Audio: MUTED (Click to enable)');
      toggleBtn.style.opacity = STATE.audioEnabled ? '1' : '0.6';
    }

    updateBtn();

    toggleBtn.addEventListener('click', () => {
      STATE.audioEnabled = !STATE.audioEnabled;
      localStorage.setItem('polyturing_audio_enabled', STATE.audioEnabled ? 'true' : 'false');
      updateBtn();
      if (STATE.audioEnabled) {
        HapticAudio.tap();
      }
    });
  }

  // ---------------------------------------------------------------------------
  // 4. SIDEBAR REAL-TIME FILTER ENGINE
  // ---------------------------------------------------------------------------
  function initSidebarFilters() {
    const filterInput = document.getElementById('sidebar-filter-input');
    const clearBtn = document.getElementById('sidebar-filter-clear');
    const treeGroups = document.querySelectorAll('.nav-tree-group');

    if (!filterInput) return;

    function applyFilter() {
      const q = filterInput.value.trim().toLowerCase();
      if (clearBtn) {
        clearBtn.style.display = q ? 'inline-block' : 'none';
      }

      treeGroups.forEach(group => {
        const titleEl = group.querySelector('.tree-title-text');
        const titleText = titleEl ? titleEl.innerText.toLowerCase() : '';
        const links = group.querySelectorAll('.nav-tree-link');
        let groupHasMatch = !q || titleText.includes(q);

        links.forEach(link => {
          const linkText = link.innerText.toLowerCase();
          const matches = !q || linkText.includes(q);
          if (matches) groupHasMatch = true;
          link.style.display = matches ? 'flex' : (q ? 'none' : 'flex');
        });

        if (groupHasMatch) {
          group.style.display = '';
          if (q) {
            group.classList.add('expanded');
          }
        } else {
          group.style.display = 'none';
        }
      });
    }

    filterInput.addEventListener('input', applyFilter);

    if (clearBtn) {
      clearBtn.addEventListener('click', () => {
        filterInput.value = '';
        applyFilter();
        filterInput.focus();
        HapticAudio.collapse();
      });
    }

    filterInput.addEventListener('keydown', (e) => {
      if (e.key === 'Escape') {
        filterInput.value = '';
        applyFilter();
        HapticAudio.collapse();
      }
    });
  }

  // ---------------------------------------------------------------------------
  // 5. SIDEBAR COLLAPSIBLE RAIL & DISTRACTION-FREE MODE
  // ---------------------------------------------------------------------------
  function initSidebarCollapseRail() {
    const dockToggle = document.getElementById('sidebar-dock-toggle');
    const closeBtn = document.getElementById('sidebar-close-btn');
    const floatingToggle = document.getElementById('floating-sidebar-toggle');

    // Restore persisted collapsed state
    if (STATE.isSidebarCollapsed && window.innerWidth > 1024) {
      document.body.classList.add('sidebar-collapsed');
    }

    function toggleRail() {
      if (window.innerWidth <= 1024) return; // Handled by mobile drawer
      const isNowCollapsed = !document.body.classList.contains('sidebar-collapsed');
      if (isNowCollapsed) {
        document.body.classList.add('sidebar-collapsed');
        HapticAudio.sidebarClose();
      } else {
        document.body.classList.remove('sidebar-collapsed');
        HapticAudio.sidebarOpen();
      }
      STATE.isSidebarCollapsed = isNowCollapsed;
      try {
        localStorage.setItem('polyturing_sidebar_collapsed', isNowCollapsed ? 'true' : 'false');
      } catch (e) {}
    }

    if (dockToggle) dockToggle.addEventListener('click', toggleRail);
    if (closeBtn) closeBtn.addEventListener('click', toggleRail);
    if (floatingToggle) floatingToggle.addEventListener('click', toggleRail);

    // Global keyboard shortcuts: ⌘B / Ctrl+B / [
    window.addEventListener('keydown', (e) => {
      const isTyping = ['INPUT', 'TEXTAREA', 'SELECT'].includes(document.activeElement?.tagName);
      if (isTyping) return;

      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'b') {
        e.preventDefault();
        toggleRail();
      } else if (e.key === '[' && !e.metaKey && !e.ctrlKey && !e.altKey) {
        e.preventDefault();
        toggleRail();
      }
    });
  }

  // ---------------------------------------------------------------------------
  // 6. KATEX MATHEMATICAL NOTATION TYPESETTING
  // ---------------------------------------------------------------------------
  function initMathRendering() {
    if (window.renderMathInElement) {
      try {
        renderMathInElement(document.body, {
          delimiters: [
            { left: '$$', right: '$$', display: true },
            { left: '$', right: '$', display: false }
          ],
          throwOnError: false
        });
      } catch (e) {
        console.warn('KaTeX auto-render warning:', e);
      }
    } else {
      setTimeout(() => {
        if (window.renderMathInElement) {
          renderMathInElement(document.body, {
            delimiters: [
              { left: '$$', right: '$$', display: true },
              { left: '$', right: '$', display: false }
            ],
            throwOnError: false
          });
        }
      }, 400);
    }
  }

  // ---------------------------------------------------------------------------
  // 7. READING PROGRESS BAR
  // ---------------------------------------------------------------------------
  function initReadingProgressBar() {
    const bar = document.getElementById('reading-progress');
    if (!bar) return;

    window.addEventListener('scroll', () => {
      const totalHeight = document.documentElement.scrollHeight - window.innerHeight;
      if (totalHeight <= 0) return;
      const progress = (window.scrollY / totalHeight) * 100;
      bar.style.width = `${Math.min(100, Math.max(0, progress))}%`;
    }, { passive: true });
  }

  // ---------------------------------------------------------------------------
  // 8. BI-DIRECTIONAL TREE NAVIGATION & EXPAND/COLLAPSE SLA
  // ---------------------------------------------------------------------------
  function initTreeNavigation() {
    const expandAllBtn = document.getElementById('expand-all-trees');
    const collapseAllBtn = document.getElementById('collapse-all-trees');

    // Expand / Collapse Tree Group Headers with Content Sync
    document.querySelectorAll('.nav-tree-header').forEach((header) => {
      header.addEventListener('click', () => {
        const group = header.closest('.nav-tree-group');
        if (!group) return;

        const isNowExpanded = !group.classList.contains('expanded');
        group.classList.toggle('expanded');

        if (isNowExpanded) {
          HapticAudio.expand();
        } else {
          HapticAudio.collapse();
        }

        // Bi-Directional Sync: Also expand/collapse corresponding chapter in content pane!
        const targetId = header.getAttribute('data-target');
        if (targetId) {
          const targetSection = document.querySelector(targetId);
          if (targetSection) {
            if (isNowExpanded) {
              targetSection.classList.remove('is-collapsed');
              if (STATE.isFocusModeActive) {
                document.querySelectorAll('.chapter-section').forEach(s => {
                  if (s !== targetSection) s.classList.add('is-collapsed');
                });
              }
              const y = targetSection.getBoundingClientRect().top + window.pageYOffset - 80;
              window.scrollTo({ top: y, behavior: 'smooth' });
              pulseElement(targetSection);
            } else {
              targetSection.classList.add('is-collapsed');
            }
          }
        }
      });
    });

    // Sub-10ms Batch Expand All / Collapse All Tree SLA
    if (expandAllBtn) {
      expandAllBtn.addEventListener('click', () => {
        STATE.isBatchExpanding = true;
        document.querySelectorAll('.nav-tree-group').forEach(g => g.classList.add('expanded'));
        document.querySelectorAll('.chapter-section').forEach(s => s.classList.remove('is-collapsed'));
        HapticAudio.expand();
        setTimeout(() => { STATE.isBatchExpanding = false; }, 50);
      });
    }

    if (collapseAllBtn) {
      collapseAllBtn.addEventListener('click', () => {
        STATE.isBatchExpanding = true;
        document.querySelectorAll('.nav-tree-group').forEach((g, idx) => {
          if (idx !== 0) g.classList.remove('expanded');
        });
        document.querySelectorAll('.chapter-section').forEach(s => s.classList.add('is-collapsed'));
        HapticAudio.collapse();
        setTimeout(() => { STATE.isBatchExpanding = false; }, 50);
      });
    }

    // When clicking any link in tree navigation, auto-expand target and smooth scroll with flash focus
    document.querySelectorAll('.nav-tree-link').forEach((link) => {
      link.addEventListener('click', (e) => {
        const href = link.getAttribute('href');
        if (href && href.startsWith('#')) {
          e.preventDefault();
          const targetEl = document.querySelector(href);
          if (targetEl) {
            ensureSectionExpanded(targetEl);
            const parentGroup = link.closest('.nav-tree-group');
            if (parentGroup) parentGroup.classList.add('expanded');

            const y = targetEl.getBoundingClientRect().top + window.pageYOffset - 80;
            window.scrollTo({ top: y, behavior: 'smooth' });

            pulseElement(targetEl);
            HapticAudio.tap();

            if (history.pushState) {
              history.pushState(null, null, href);
            }
          }
        }
      });
    });

    // Mobile drawer quick tool links
    document.querySelectorAll('.mobile-tool-link').forEach((link) => {
      link.addEventListener('click', (e) => {
        const href = link.getAttribute('href');
        if (href && href.startsWith('#')) {
          e.preventDefault();
          const targetEl = document.querySelector(href);
          if (targetEl) {
            ensureSectionExpanded(targetEl);
            const y = targetEl.getBoundingClientRect().top + window.pageYOffset - 80;
            window.scrollTo({ top: y, behavior: 'smooth' });
            pulseElement(targetEl);
            HapticAudio.tap();
            if (history.pushState) {
              history.pushState(null, null, href);
            }
          }
        }
      });
    });
  }

  // ---------------------------------------------------------------------------
  // 9. DYNAMIC DRAG-TO-RESIZE SIDEBAR SPLITTER
  // ---------------------------------------------------------------------------
  function initSidebarResizer() {
    const resizer = document.getElementById('sidebar-resizer');
    const sidebar = document.getElementById('main-sidebar');
    if (!resizer || !sidebar) return;

    const MIN_WIDTH = 240;
    const MAX_WIDTH = 560;
    const DEFAULT_WIDTH = 320;

    // 1. Restore persisted width from localStorage
    try {
      const savedWidth = localStorage.getItem('polyturing_sidebar_width');
      if (savedWidth) {
        const parsed = parseInt(savedWidth, 10);
        if (!isNaN(parsed) && parsed >= MIN_WIDTH && parsed <= MAX_WIDTH) {
          document.documentElement.style.setProperty('--sidebar-width', `${parsed}px`);
        }
      }
    } catch (e) {
      console.warn('Could not read sidebar width from localStorage:', e);
    }

    let isDragging = false;
    let startX = 0;
    let startWidth = 0;

    function onPointerDown(e) {
      if (e.type === 'mousedown' && e.button !== 0) return;
      if (document.body.classList.contains('sidebar-collapsed')) return;
      if (window.innerWidth <= 1024) return;

      isDragging = true;
      startX = e.type.startsWith('touch') ? e.touches[0].clientX : e.clientX;
      startWidth = sidebar.getBoundingClientRect().width;

      resizer.classList.add('is-resizing');
      document.body.classList.add('resizing-sidebar');

      window.addEventListener('mousemove', onPointerMove, { passive: false });
      window.addEventListener('touchmove', onPointerMove, { passive: false });
      window.addEventListener('mouseup', onPointerUp);
      window.addEventListener('touchend', onPointerUp);
    }

    function onPointerMove(e) {
      if (!isDragging) return;
      e.preventDefault();

      const currentX = e.type.startsWith('touch') ? e.touches[0].clientX : e.clientX;
      const deltaX = currentX - startX;
      let newWidth = Math.round(startWidth + deltaX);

      if (newWidth < MIN_WIDTH) newWidth = MIN_WIDTH;
      if (newWidth > MAX_WIDTH) newWidth = MAX_WIDTH;

      document.documentElement.style.setProperty('--sidebar-width', `${newWidth}px`);
    }

    function onPointerUp() {
      if (!isDragging) return;
      isDragging = false;

      resizer.classList.remove('is-resizing');
      document.body.classList.remove('resizing-sidebar');

      window.removeEventListener('mousemove', onPointerMove);
      window.removeEventListener('touchmove', onPointerMove);
      window.removeEventListener('mouseup', onPointerUp);
      window.removeEventListener('touchend', onPointerUp);

      const styleVal = parseInt(document.documentElement.style.getPropertyValue('--sidebar-width'), 10);
      const finalWidth = !isNaN(styleVal) ? styleVal : Math.round(sidebar.getBoundingClientRect().width);
      try {
        localStorage.setItem('polyturing_sidebar_width', finalWidth);
      } catch (e) {
        console.warn('Could not save sidebar width to localStorage:', e);
      }
    }

    resizer.addEventListener('mousedown', onPointerDown);
    resizer.addEventListener('touchstart', onPointerDown, { passive: true });

    resizer.addEventListener('dblclick', () => {
      document.documentElement.style.setProperty('--sidebar-width', `${DEFAULT_WIDTH}px`);
      try {
        localStorage.setItem('polyturing_sidebar_width', DEFAULT_WIDTH);
        HapticAudio.tap();
      } catch (e) {
        console.warn('Could not reset sidebar width in localStorage:', e);
      }
    });
  }

  // ---------------------------------------------------------------------------
  // 10. YEAR 2500 MOBILE SLIDE-OUT NAVIGATION DRAWER
  // ---------------------------------------------------------------------------
  function initMobileDrawer() {
    const toggleBtn = document.getElementById('mobile-nav-toggle');
    const closeBtn = document.getElementById('mobile-drawer-close');
    const backdrop = document.getElementById('mobile-drawer-backdrop');
    const sidebar = document.getElementById('main-sidebar');

    if (!sidebar) return;

    function openDrawer() {
      sidebar.classList.add('mobile-open');
      if (backdrop) backdrop.classList.add('show');
      document.body.style.overflow = 'hidden';
      HapticAudio.sidebarOpen();
    }

    function closeDrawer() {
      sidebar.classList.remove('mobile-open');
      if (backdrop) backdrop.classList.remove('show');
      document.body.style.overflow = '';
      HapticAudio.sidebarClose();
    }

    if (toggleBtn) {
      toggleBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        openDrawer();
      });
    }

    if (closeBtn) {
      closeBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        closeDrawer();
      });
    }

    if (backdrop) {
      backdrop.addEventListener('click', closeDrawer);
    }

    window.addEventListener('keydown', (e) => {
      if (e.key === 'Escape' && sidebar.classList.contains('mobile-open')) {
        closeDrawer();
      }
    });

    sidebar.querySelectorAll('.nav-tree-link, .mobile-tool-link').forEach((link) => {
      link.addEventListener('click', () => {
        if (window.innerWidth <= 1024) {
          closeDrawer();
        }
      });
    });
  }

  // ---------------------------------------------------------------------------
  // 11. COLLAPSIBLE CHAPTERS & ACCORDION / FOCUS MODES
  // ---------------------------------------------------------------------------
  function initCollapsibleChapters() {
    const expandAllBtn = document.getElementById('mode-expand-all');
    const focusModeBtn = document.getElementById('mode-focus-accordion');
    const collapseAllBtn = document.getElementById('mode-collapse-all');

    document.querySelectorAll('.chapter-header-trigger').forEach((trigger) => {
      trigger.addEventListener('click', () => {
        const section = trigger.closest('.chapter-section');
        if (!section) return;

        const willExpand = section.classList.contains('is-collapsed');

        if (willExpand) {
          if (STATE.isFocusModeActive) {
            document.querySelectorAll('.chapter-section').forEach(s => {
              if (s !== section) s.classList.add('is-collapsed');
            });
          }
          section.classList.remove('is-collapsed');
          syncNavGroupForSection(section.id, true);
          HapticAudio.expand();
        } else {
          section.classList.add('is-collapsed');
          HapticAudio.collapse();
        }
      });
    });

    if (expandAllBtn) {
      expandAllBtn.addEventListener('click', () => {
        STATE.isFocusModeActive = false;
        setActiveModeBtn(expandAllBtn);
        STATE.isBatchExpanding = true;
        document.querySelectorAll('.chapter-section').forEach(s => s.classList.remove('is-collapsed'));
        HapticAudio.expand();
        setTimeout(() => { STATE.isBatchExpanding = false; }, 50);
      });
    }

    if (focusModeBtn) {
      focusModeBtn.addEventListener('click', () => {
        STATE.isFocusModeActive = true;
        setActiveModeBtn(focusModeBtn);
        STATE.isBatchExpanding = true;
        const sections = Array.from(document.querySelectorAll('.chapter-section'));
        if (sections.length > 0) {
          sections[0].classList.remove('is-collapsed');
          for (let i = 1; i < sections.length; i++) {
            sections[i].classList.add('is-collapsed');
          }
        }
        HapticAudio.tap();
        setTimeout(() => { STATE.isBatchExpanding = false; }, 50);
      });
    }

    if (collapseAllBtn) {
      collapseAllBtn.addEventListener('click', () => {
        STATE.isFocusModeActive = false;
        setActiveModeBtn(collapseAllBtn);
        STATE.isBatchExpanding = true;
        document.querySelectorAll('.chapter-section').forEach(s => s.classList.add('is-collapsed'));
        HapticAudio.collapse();
        setTimeout(() => { STATE.isBatchExpanding = false; }, 50);
      });
    }

    function setActiveModeBtn(activeBtn) {
      [expandAllBtn, focusModeBtn, collapseAllBtn].forEach(btn => {
        if (btn) btn.classList.remove('active');
      });
      if (activeBtn) activeBtn.classList.add('active');
    }
  }

  function ensureSectionExpanded(targetElement) {
    const parentSection = targetElement.closest('.chapter-section') || targetElement;
    if (parentSection) {
      if (parentSection.classList.contains('is-collapsed')) {
        if (STATE.isFocusModeActive) {
          document.querySelectorAll('.chapter-section').forEach(s => {
            if (s !== parentSection) s.classList.add('is-collapsed');
          });
        }
        parentSection.classList.remove('is-collapsed');
      }
      syncNavGroupForSection(parentSection.id, true);
    }
  }

  function syncNavGroupForSection(sectionId, expand) {
    const treeLink = document.querySelector(`.nav-tree-link[href="#${sectionId}"]`) ||
                     document.querySelector(`.nav-tree-header[data-target="#${sectionId}"]`);
    if (treeLink) {
      const group = treeLink.closest('.nav-tree-group');
      if (group && expand) {
        group.classList.add('expanded');
      }
    }
  }

  function pulseElement(el) {
    const targetCard = el.classList.contains('glass-card') || el.classList.contains('chapter-section')
      ? el 
      : el.closest('.glass-card, .canvas-container, .callout, .chapter-section') || el;
    
    targetCard.classList.remove('flash-focus');
    void targetCard.offsetWidth;
    targetCard.classList.add('flash-focus');
    setTimeout(() => targetCard.classList.remove('flash-focus'), 1900);
  }

  // ---------------------------------------------------------------------------
  // 12. FULLSCREEN VISUAL THEATER WORKBENCH (<dialog> & Sub-ms Zoom/Pan)
  // ---------------------------------------------------------------------------
  function initCanvasZoomAndTheater() {
    const theaterModal = document.getElementById('theater-modal');
    const theaterTitle = document.getElementById('theater-title');
    const theaterSurface = document.getElementById('theater-surface');
    const theaterViewport = document.getElementById('theater-viewport');
    const closeTheaterBtn = document.getElementById('close-theater-btn');
    const theaterZoomIn = document.getElementById('theater-zoom-in');
    const theaterZoomOut = document.getElementById('theater-zoom-out');
    const theaterZoomReset = document.getElementById('theater-zoom-reset');
    const theaterZoomFit = document.getElementById('theater-zoom-fit');
    const theaterFullscreen = document.getElementById('theater-fullscreen');
    const theaterZoomBadge = document.getElementById('theater-zoom-badge');

    let activeTheaterZoom = 1.0;
    let activeTheaterPanX = 0;
    let activeTheaterPanY = 0;
    let lastActiveSourceCanvas = null;
    let savedScrollY = 0;

    // Helper: calculate optimal viewport fit scale
    function calculateFitScale(svgEl) {
      if (!theaterViewport || !svgEl) return 1.0;
      let vbW = 0, vbH = 0;
      const vb = svgEl.getAttribute('viewBox');
      if (vb) {
        const parts = vb.trim().split(/[\s,]+/).map(Number);
        if (parts.length >= 4 && parts[2] > 0 && parts[3] > 0) {
          vbW = parts[2];
          vbH = parts[3];
        }
      }
      if (!vbW || !vbH) {
        if (svgEl.viewBox && svgEl.viewBox.baseVal && svgEl.viewBox.baseVal.width > 0) {
          vbW = svgEl.viewBox.baseVal.width;
          vbH = svgEl.viewBox.baseVal.height;
        } else {
          const rect = svgEl.getBoundingClientRect();
          vbW = rect.width || 800;
          vbH = rect.height || 600;
        }
      }
      vbW = vbW || 800;
      vbH = vbH || 500;
      const availW = Math.max(300, (theaterViewport.clientWidth || window.innerWidth) - 64);
      const availH = Math.max(200, (theaterViewport.clientHeight || window.innerHeight) - 64);
      const fitW = availW / vbW;
      const fitH = availH / vbH;
      const scale = Math.min(fitW, fitH);
      return Math.max(0.15, Math.min(scale, 3.5));
    }

    function applyTheaterTransform() {
      if (!theaterSurface) return;
      theaterSurface.style.transform = `translate(${activeTheaterPanX}px, ${activeTheaterPanY}px) scale(${activeTheaterZoom})`;
      if (theaterZoomBadge) {
        theaterZoomBadge.innerText = `${Math.round(activeTheaterZoom * 100)}%`;
      }
    }

    // Initialize all canvas containers
    document.querySelectorAll('.canvas-container').forEach((canvas) => {
      const viewport = canvas.querySelector('.canvas-viewport');
      const surface = canvas.querySelector('.canvas-content-surface');
      const zoomInBtn = canvas.querySelector('.zoom-in-btn');
      const zoomOutBtn = canvas.querySelector('.zoom-out-btn');
      const zoomResetBtn = canvas.querySelector('.zoom-reset-btn');
      const zoomBadge = canvas.querySelector('.zoom-level-badge');
      const fullscreenBtn = canvas.querySelector('.fullscreen-btn');
      const titleTag = canvas.querySelector('.canvas-type-tag')?.innerText || 'Architecture View';

      if (!viewport || !surface) return;

      let zoom = 1.0;
      let panX = 0;
      let panY = 0;
      let isPanning = false;
      let startX = 0;
      let startY = 0;

      function applyCanvasTransform() {
        surface.style.transform = `scale(${zoom}) translate(${panX}px, ${panY}px)`;
        if (zoomBadge) zoomBadge.innerText = `${Math.round(zoom * 100)}%`;
      }

      if (zoomInBtn) {
        zoomInBtn.addEventListener('click', () => {
          zoom = Math.min(3.5, zoom + 0.25);
          applyCanvasTransform();
          HapticAudio.tap();
        });
      }

      if (zoomOutBtn) {
        zoomOutBtn.addEventListener('click', () => {
          zoom = Math.max(0.5, zoom - 0.25);
          applyCanvasTransform();
          HapticAudio.tap();
        });
      }

      if (zoomResetBtn) {
        zoomResetBtn.addEventListener('click', () => {
          zoom = 1.0;
          panX = 0;
          panY = 0;
          applyCanvasTransform();
          HapticAudio.tap();
        });
      }

      viewport.addEventListener('mousedown', (e) => {
        if (e.target.closest('button')) return;
        isPanning = true;
        startX = e.clientX - panX;
        startY = e.clientY - panY;
        viewport.style.cursor = 'grabbing';
      });

      window.addEventListener('mousemove', (e) => {
        if (!isPanning) return;
        panX = e.clientX - startX;
        panY = e.clientY - startY;
        applyCanvasTransform();
      });

      window.addEventListener('mouseup', () => {
        if (isPanning) {
          isPanning = false;
          viewport.style.cursor = 'grab';
        }
      });

      viewport.addEventListener('wheel', (e) => {
        if (e.ctrlKey || e.metaKey) {
          e.preventDefault();
          const delta = e.deltaY < 0 ? 0.15 : -0.15;
          zoom = Math.min(3.5, Math.max(0.5, zoom + delta));
          applyCanvasTransform();
        }
      }, { passive: false });

      if (fullscreenBtn && theaterModal && theaterSurface) {
        fullscreenBtn.addEventListener('click', () => {
          openTheater(titleTag, canvas);
        });
      }
    });

    // Theater Mode Opening & Lifecycle
    function openTheater(title, sourceCanvas) {
      if (!theaterModal || !theaterSurface || !sourceCanvas) return;

      // 1. Save scroll position precisely BEFORE opening
      savedScrollY = window.scrollY || window.pageYOffset;
      lastActiveSourceCanvas = sourceCanvas;

      const titleText = theaterTitle?.querySelector('.theater-title-text');
      if (titleText) {
        titleText.innerText = title;
      } else if (theaterTitle) {
        theaterTitle.innerText = title;
      }

      theaterSurface.innerHTML = '';

      // 2. Clone SVG with unconstrained viewBox pixel dimensions
      let originalSvg = sourceCanvas.querySelector('.mermaid svg') || sourceCanvas.querySelector('svg');
      if (originalSvg) {
        const cleanSvg = originalSvg.cloneNode(true);
        cleanSvg.removeAttribute('style');

        let vbW = 0, vbH = 0;
        const vb = cleanSvg.getAttribute('viewBox');
        if (vb) {
          const parts = vb.trim().split(/[\s,]+/).map(Number);
          if (parts.length >= 4 && parts[2] > 0 && parts[3] > 0) {
            vbW = parts[2];
            vbH = parts[3];
          }
        }
        if (vbW && vbH) {
          cleanSvg.style.width = vbW + 'px';
          cleanSvg.style.height = vbH + 'px';
        } else {
          cleanSvg.style.width = '100%';
          cleanSvg.style.height = 'auto';
        }
        cleanSvg.style.maxWidth = 'none';
        cleanSvg.style.maxHeight = 'none';
        theaterSurface.appendChild(cleanSvg);
      } else {
        const surf = sourceCanvas.querySelector('.canvas-content-surface');
        if (surf) theaterSurface.innerHTML = surf.innerHTML;
      }

      // 3. Open modal dialog
      if (typeof theaterModal.showModal === 'function') {
        try {
          theaterModal.showModal();
        } catch (e) {
          theaterModal.classList.add('theater-open');
        }
      } else {
        theaterModal.classList.add('theater-open');
      }

      document.body.classList.add('theater-open');
      document.body.style.overflow = 'hidden';

      // 4. Auto-fit scale on open
      const renderedSvg = theaterSurface.querySelector('svg');
      activeTheaterZoom = calculateFitScale(renderedSvg);
      activeTheaterPanX = 0;
      activeTheaterPanY = 0;
      applyTheaterTransform();

      HapticAudio.theaterOpen();
    }

    function closeTheater() {
      if (!theaterModal) return;

      if (typeof theaterModal.close === 'function') {
        try {
          theaterModal.close();
        } catch (e) {
          theaterModal.classList.remove('theater-open');
        }
      } else {
        theaterModal.classList.remove('theater-open');
      }

      document.body.classList.remove('theater-open');
      document.body.style.overflow = '';
      if (theaterSurface) theaterSurface.innerHTML = '';

      if (document.fullscreenElement) {
        document.exitFullscreen().catch(() => {});
      }

      // Restore user's exact scroll position instantly (zero jump / lost context)
      window.scrollTo({ top: savedScrollY, behavior: 'instant' });

      // Highlight source container with flash-focus glow
      if (lastActiveSourceCanvas) {
        const container = lastActiveSourceCanvas;
        requestAnimationFrame(() => {
          container.classList.add('flash-focus');
          setTimeout(() => container?.classList.remove('flash-focus'), 1800);
        });
      }

      HapticAudio.sidebarClose();
    }

    if (closeTheaterBtn) closeTheaterBtn.addEventListener('click', closeTheater);

    if (theaterZoomIn) {
      theaterZoomIn.addEventListener('click', () => {
        activeTheaterZoom = Math.min(4.0, activeTheaterZoom + 0.15);
        applyTheaterTransform();
        HapticAudio.tap();
      });
    }

    if (theaterZoomOut) {
      theaterZoomOut.addEventListener('click', () => {
        activeTheaterZoom = Math.max(0.25, activeTheaterZoom - 0.15);
        applyTheaterTransform();
        HapticAudio.tap();
      });
    }

    if (theaterZoomReset) {
      theaterZoomReset.addEventListener('click', () => {
        activeTheaterZoom = 1.0;
        activeTheaterPanX = 0;
        activeTheaterPanY = 0;
        applyTheaterTransform();
        HapticAudio.tap();
      });
    }

    if (theaterZoomFit) {
      theaterZoomFit.addEventListener('click', () => {
        const svg = theaterSurface?.querySelector('svg');
        if (svg) {
          activeTheaterZoom = calculateFitScale(svg);
        } else {
          activeTheaterZoom = 1.0;
        }
        activeTheaterPanX = 0;
        activeTheaterPanY = 0;
        applyTheaterTransform();
        HapticAudio.tap();
      });
    }

    if (theaterFullscreen) {
      theaterFullscreen.addEventListener('click', () => {
        if (!document.fullscreenElement) {
          theaterModal?.requestFullscreen?.().catch(() => {});
        } else {
          document.exitFullscreen?.().catch(() => {});
        }
        HapticAudio.tap();
      });
    }

    // Theater drag panning
    let isTheaterPanning = false;
    let theaterStartX = 0;
    let theaterStartY = 0;

    if (theaterViewport) {
      theaterViewport.addEventListener('mousedown', (e) => {
        if (e.target.closest('button')) return;
        // Don't pan if user is highlighting text inside SVG
        if (['text', 'tspan', 'span'].includes(e.target.tagName.toLowerCase())) return;

        isTheaterPanning = true;
        theaterStartX = e.clientX - activeTheaterPanX;
        theaterStartY = e.clientY - activeTheaterPanY;
        theaterViewport.style.cursor = 'grabbing';
      });

      window.addEventListener('mousemove', (e) => {
        if (!isTheaterPanning) return;
        activeTheaterPanX = e.clientX - theaterStartX;
        activeTheaterPanY = e.clientY - theaterStartY;
        applyTheaterTransform();
      });

      window.addEventListener('mouseup', () => {
        if (isTheaterPanning) {
          isTheaterPanning = false;
          if (theaterViewport) theaterViewport.style.cursor = 'grab';
        }
      });

      theaterViewport.addEventListener('wheel', (e) => {
        e.preventDefault();
        const factor = e.deltaY < 0 ? 1.15 : 0.87;
        activeTheaterZoom = Math.max(0.15, Math.min(5.0, Number((activeTheaterZoom * factor).toFixed(2))));
        applyTheaterTransform();
      }, { passive: false });
    }

    // Escape closes theater
    window.addEventListener('keydown', (e) => {
      if (e.key === 'Escape') {
        const isOpen = theaterModal?.hasAttribute('open') || theaterModal?.classList.contains('theater-open');
        if (isOpen) {
          closeTheater();
        }
      }
    });
  }

  // ---------------------------------------------------------------------------
  // 13. DYNAMIC SCROLLSPY, AMBIENT PROGRESS & BACK-TO-TOP FAB ENGINE
  // ---------------------------------------------------------------------------
  function initScrollSpy() {
    const navTreeLinks = Array.from(document.querySelectorAll('.nav-tree-link'));
    const navTreeGroups = Array.from(document.querySelectorAll('.nav-tree-group'));
    const treeHeaders = Array.from(document.querySelectorAll('.nav-tree-header'));
    const globalProgressBar = document.getElementById('global-reading-progress');
    const backToTopFab = document.getElementById('back-to-top-fab');

    // Collect all unique targets referenced in left nav
    const targetMap = new Map();
    navTreeLinks.forEach(link => {
      const href = link.getAttribute('href');
      if (href && href.startsWith('#')) {
        const el = document.querySelector(href);
        if (el) targetMap.set(href.slice(1), el);
      }
    });

    const trackedItems = Array.from(targetMap.entries()).map(([id, element]) => ({ id, element }));

    let isTicking = false;
    let currentActiveId = null;

    window.addEventListener('scroll', () => {
      if (STATE.isNavClicking) return;
      if (!isTicking) {
        window.requestAnimationFrame(() => {
          updateActiveNav();
          isTicking = false;
        });
        isTicking = true;
      }
    }, { passive: true });

    function updateActiveNav() {
      // 1. Update ambient reading progress bar
      const totalScrollable = document.documentElement.scrollHeight - window.innerHeight;
      if (totalScrollable > 0 && globalProgressBar) {
        const pct = Math.min(100, Math.max(0, Math.round((window.scrollY / totalScrollable) * 100)));
        globalProgressBar.style.width = `${pct}%`;
      }

      // 2. Manage Floating Back-to-Top FAB visibility
      if (backToTopFab) {
        if (window.scrollY > 300) {
          backToTopFab.classList.add('visible');
        } else {
          backToTopFab.classList.remove('visible');
        }
      }

      const scrollPos = window.scrollY + 120; // Header offset
      const visibleItems = trackedItems.filter(item => {
        return item.element.offsetHeight > 0 || item.element.offsetParent !== null;
      });

      if (visibleItems.length === 0) return;

      let activeId = visibleItems[0].id;
      const isAtBottom = (window.innerHeight + window.scrollY) >= (document.documentElement.scrollHeight - 70);

      if (isAtBottom) {
        activeId = visibleItems[visibleItems.length - 1].id;
      } else {
        for (let i = 0; i < visibleItems.length; i++) {
          const item = visibleItems[i];
          const top = item.element.getBoundingClientRect().top + window.scrollY;
          if (scrollPos >= top) {
            activeId = item.id;
          } else {
            break;
          }
        }
      }

      if (activeId && activeId !== currentActiveId) {
        currentActiveId = activeId;
        applyActiveNavId(activeId);
      }
    }

    function applyActiveNavId(id) {
      let matchedGroup = null;

      // 1. Left Navigation Active Classes
      navTreeLinks.forEach(link => {
        if (link.getAttribute('href') === `#${id}`) {
          link.classList.add('active');
          const parentGroup = link.closest('.nav-tree-group');
          if (parentGroup) {
            matchedGroup = parentGroup;
            treeHeaders.forEach(h => h.classList.remove('active-branch'));
            parentGroup.querySelector('.nav-tree-header')?.classList.add('active-branch');
          }
        } else {
          link.classList.remove('active');
        }
      });

      // 2. Auto-expand active group in sidebar if collapsed & not in batch operation
      if (matchedGroup && !matchedGroup.classList.contains('expanded') && !STATE.isBatchExpanding && !STATE.isNavClicking) {
        matchedGroup.classList.add('expanded');
        matchedGroup.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
      }
    }

    // Back to top FAB action
    if (backToTopFab) {
      backToTopFab.addEventListener('click', () => {
        window.scrollTo({ top: 0, behavior: 'smooth' });
        HapticAudio.play('tap');
      });
    }

    // Initial calculation
    updateActiveNav();
  }

  // ---------------------------------------------------------------------------
  // 14. COPY CODE SNIPPET BUTTONS WITH CELEBRATION CHIME
  // ---------------------------------------------------------------------------
  function initCopyButtons() {
    document.querySelectorAll('.copy-btn').forEach((btn) => {
      btn.addEventListener('click', async () => {
        const codeId = btn.getAttribute('data-target');
        let textToCopy = '';
        if (codeId) {
          const target = document.getElementById(codeId);
          if (target) textToCopy = target.innerText;
        } else {
          const codeBlock = btn.closest('.code-container')?.querySelector('code');
          if (codeBlock) textToCopy = codeBlock.innerText;
        }

        if (textToCopy) {
          try {
            await navigator.clipboard.writeText(textToCopy);
            const origHtml = btn.innerHTML;
            btn.innerHTML = `
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg>
              <span>Copied!</span>
            `;
            btn.classList.add('copied');
            HapticAudio.copySuccess();

            setTimeout(() => {
              btn.innerHTML = origHtml;
              btn.classList.remove('copied');
            }, 2000);
          } catch (err) {
            console.error('Copy failed', err);
          }
        }
      });
    });
  }

  // ---------------------------------------------------------------------------
  // 15. GLOBAL SEARCH MODAL & INDEX
  // ---------------------------------------------------------------------------
  const SEARCH_DOCS = [
    { id: 'ch0-manifesto', title: 'Platform Manifesto: The Deterministic Intelligence Substrate for AI Coding Assistants', snippet: 'Bridging raw LLM reasoning and real-world codebases with deterministic AST knowledge.', category: 'Manifesto' },
    { id: 'sec-context-crisis', title: 'The Context Window Crisis & The Myth of Raw Grep', snippet: 'Context window saturation, token exhaustion, blind file grepping, and code hallucinations.', category: 'Manifesto' },
    { id: 'sec-polyturing-triad', title: 'The PolyTuring Triad: Omni-Graph, Session Explorer, Living Docs', snippet: 'Deterministic AST substrate, trajectory intelligence cockpit, and architectural contract engine.', category: 'Manifesto' },
    { id: 'sec-antigravity-native', title: 'Antigravity-Native Substrate & Multi-Assistant Roadmap', snippet: 'Built natively for Google Antigravity (agy, brain, hooks) with extensibility for Copilot & Claude Code.', category: 'Manifesto' },
    { id: 'ch1-omni-graph', title: 'Omni-Graph: 100% Local Semantic Knowledge Hub', snippet: 'AST graph algebra, 384-dim TEI vector embeddings, HNSW index, SurrealDB v2.', category: 'Omni-Graph' },
    { id: 'sec-omni-arch', title: 'Omni-Graph: 4-Tier Container Architecture', snippet: 'SurrealDB, HuggingFace TEI, Rust Axum Orchestrator, Cosmograph WebGL Visualizer.', category: 'Omni-Graph' },
    { id: 'sec-omni-galaxies', title: 'Galaxy Subsystems & Louvain / Leiden Modularity Clustering', snippet: 'Scale-free code networks, Modularity Q optimization, edge-weighted affinity, Leiden guarantees.', category: 'Omni-Graph' },
    { id: 'sec-omni-coupling', title: 'Robert C. Martin Coupling Metrics (Ca, Ce, I) & Blast Radius Contracts', snippet: 'Afferent/Efferent coupling, Instability metric, Core Foundation vs Leaf roles, boundary contracts.', category: 'Omni-Graph' },
    { id: 'sec-omni-ebf', title: 'Ephemeral Branch Fabric (EBF) & Git Worktrees', snippet: 'O(1) worktree peeking, Prior LPA seed inheritance, zero-storm rebase suppression.', category: 'Omni-Graph' },
    { id: 'sec-omni-ast', title: 'Polyglot AST Grammar & Struct Field Extraction', snippet: 'Rust fields/impls/macros, Go methods/structs/fields, TS arrow fns/interfaces/properties.', category: 'Omni-Graph' },
    { id: 'sec-omni-budget', title: 'Two-Tier Context Condenser Token Budget Math (≤ 1500 tokens)', snippet: 'Strict character bounding: 3,800 chars for definitions, ≥ 1,800 chars for call traces.', category: 'Omni-Graph' },
    { id: 'sec-omni-relational', title: 'SurrealDB Relational Indexing & Zero-Blackout Re-indexing', snippet: 'idx_edge_in/out indexes, atomic vector generation before node pruning.', category: 'Omni-Graph' },
    { id: 'token-simulator', title: 'Interactive Context Condenser Token Simulator', snippet: 'Simulate real-time token counts, root snippets, neighbor caps, and SLA compliance.', category: 'Interactive' },
    { id: 'api-playground', title: 'Omni-Graph Live API Playground & Curl Generator', snippet: '/api/symbol, /api/references, /api/condense, /api/query, /api/workspaces, /api/worktrees.', category: 'Interactive' },
    { id: 'ch2-session-explorer', title: 'Session Explorer: Agent Observability & Telemetry', snippet: 'Fast JSONL parser, prompt indexing, tool invocation insights, SQLite cache, Go binary.', category: 'Session Explorer' },
    { id: 'sec-session-arch', title: 'Session Explorer: Go + Bun Standalone Binary Architecture', snippet: 'Go backend, Bun frontend, marked.js markdown engine, single standalone binary.', category: 'Session Explorer' },
    { id: 'sec-session-wip', title: 'Agent Work-in-Progress (WIP) Recovery & Disambiguation', snippet: 'Instant context recovery from brain logs, mandatory halt on ambiguous sessions, /most-recent-wip.', category: 'Session Explorer' },
    { id: 'sec-session-viewer', title: 'In-Browser Code Viewer: 22px Gutter Invariant & Persistent In-Page Find', snippet: 'Zero dual-scroll, 0px line drift across 10k lines, floating sticky find widget with count.', category: 'Session Explorer' },
    { id: 'sec-agent-context', title: 'Universal Agent Context & Behavioral Matrix (R-039)', snippet: 'Dynamic transcript extraction, 5-tier scope precedence cascade, active vs on-demand skills.', category: 'Session Explorer' },
    { id: 'sec-session-swagger', title: 'Embedded OpenAPI 3.0 & Swagger UI', snippet: 'Offline API explorer at /docs, raw OpenAPI JSON contract at /api/openapi.json.', category: 'Session Explorer' },
    { id: 'sec-session-usage', title: 'Session Explorer: Telemetry & CLI Quickstart', snippet: 'make session-explorer, make run PORT=9876, --no-open headless mode, export markdown.', category: 'Session Explorer' },
    { id: 'ch3-assistant-matrix', title: 'Universal Coding Assistant Matrix & Extensibility Blueprint', snippet: 'Native Google Antigravity vs. GitHub Copilot, Claude Code, and Cursor roadmap.', category: 'Ecosystem' },
    { id: 'sec-assistant-roadmap', title: 'Multi-Assistant Phasing: Antigravity, Copilot, Claude Code, Cursor', snippet: 'Three-phase rollout: Native substrate, Model Context Protocol (MCP), and Copilot LSP sidecars.', category: 'Ecosystem' },
    { id: 'sec-assistant-mcp', title: 'Universal Model Context Protocol (MCP) Architecture', snippet: 'MCP server bridge connecting any coding assistant to local AST knowledge and session telemetry.', category: 'Ecosystem' },
    { id: 'sec-appendix-rigor', title: 'Engineering Rigor, SDD Constitution & Verification Gates', snippet: 'Spec-Driven Development, Reflexion loops, 3-tier reconnaissance ladder, CI/CD gates.', category: 'Governance' },
    { id: 'sec-sdd-governance', title: 'Spec-Driven Development (SDD) & SDLC Finite State Machine', snippet: 'Contract before code, numbered requirements (R-XXX), status verification gates.', category: 'Governance' },
    { id: 'sec-agent-shield', title: 'The Agentic Defense Shield & Reconnaissance Gates', snippet: '3-tier recon ladder, pre-invocation hook, 150-line gate, recon markers.', category: 'Governance' },
    { id: 'sec-ci-gates', title: 'Continuous Integration Matrix & Polyglot Test Runners', snippet: 'ci.yml, tool-ci.yml, docs-publish.yml, polyglot test runners, specs validation.', category: 'Governance' },
    { id: 'sec-docs-drift', title: 'The Documentation Synchronization Gate & Rule 06 Verification', snippet: '4-point documentation gate, zero drift policy, Rule 06 standards.', category: 'Governance' }
  ];

  function initSearchModal() {
    const modal = document.getElementById('search-modal');
    const trigger = document.getElementById('search-trigger');
    const input = document.getElementById('search-input');
    const resultsContainer = document.getElementById('search-results');

    if (!modal || !trigger || !input || !resultsContainer) return;

    function openModal() {
      modal.classList.add('open');
      input.value = '';
      renderResults('');
      setTimeout(() => input.focus(), 50);
      HapticAudio.searchOpen();
    }

    function closeModal() {
      modal.classList.remove('open');
      HapticAudio.sidebarClose();
    }

    trigger.addEventListener('click', openModal);

    document.addEventListener('keydown', (e) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        modal.classList.contains('open') ? closeModal() : openModal();
      } else if (e.key === 'Escape' && modal.classList.contains('open')) {
        closeModal();
      } else if (e.key === '/' && !['INPUT', 'TEXTAREA'].includes(document.activeElement?.tagName)) {
        e.preventDefault();
        openModal();
      }
    });

    modal.addEventListener('click', (e) => {
      if (e.target === modal) closeModal();
    });

    input.addEventListener('input', (e) => {
      renderResults(e.target.value.trim().toLowerCase());
    });

    function renderResults(q) {
      resultsContainer.innerHTML = '';
      const filtered = q === '' 
        ? SEARCH_DOCS.slice(0, 6)
        : SEARCH_DOCS.filter(d => 
            d.title.toLowerCase().includes(q) || 
            d.snippet.toLowerCase().includes(q) ||
            d.category.toLowerCase().includes(q)
          );

      if (filtered.length === 0) {
        resultsContainer.innerHTML = '<div style="padding: 24px; text-align: center; color: var(--text-dim);">No matching documentation found</div>';
        return;
      }

      filtered.forEach((item, idx) => {
        const a = document.createElement('a');
        a.className = `search-result-item ${idx === 0 ? 'selected' : ''}`;
        a.href = `#${item.id}`;
        a.innerHTML = `
          <div style="display: flex; justify-content: space-between; align-items: baseline;">
            <div class="result-title">${item.title}</div>
            <span style="font-size: 0.72rem; color: var(--accent-cyan); text-transform: uppercase;">${item.category}</span>
          </div>
          <div class="result-snippet">${item.snippet}</div>
        `;
        a.addEventListener('click', (e) => {
          closeModal();
          const targetSec = document.getElementById(item.id);
          if (targetSec) {
            ensureSectionExpanded(targetSec);
            const y = targetSec.getBoundingClientRect().top + window.pageYOffset - 80;
            window.scrollTo({ top: y, behavior: 'smooth' });
            pulseElement(targetSec);
            HapticAudio.tap();
          }
        });
        resultsContainer.appendChild(a);
      });
    }
  }

  // ---------------------------------------------------------------------------
  // 16. TOKEN BUDGET SIMULATOR
  // ---------------------------------------------------------------------------
  function initTokenSimulator() {
    const rootSlider = document.getElementById('sim-root-chars');
    const neighborSlider = document.getElementById('sim-neighbor-chars');
    const callsSlider = document.getElementById('sim-calls-chars');
    const valRoot = document.getElementById('val-root-chars');
    const valNeighbor = document.getElementById('val-neighbor-chars');
    const valCalls = document.getElementById('val-calls-chars');
    const totalCharsEl = document.getElementById('sim-total-chars');
    const totalTokensEl = document.getElementById('sim-total-tokens');
    const slaStatusEl = document.getElementById('sim-sla-status');
    const barFill = document.getElementById('sim-bar-fill');

    if (!rootSlider || !neighborSlider || !callsSlider) return;

    function recalculate() {
      const r = parseInt(rootSlider.value, 10);
      const n = parseInt(neighborSlider.value, 10);
      const c = parseInt(callsSlider.value, 10);

      if (valRoot) valRoot.innerText = `${r} chars`;
      if (valNeighbor) valNeighbor.innerText = `${n} chars`;
      if (valCalls) valCalls.innerText = `${c} chars`;

      const totalChars = r + n + c;
      const totalTokens = Math.round(totalChars / 4.0);

      if (totalCharsEl) totalCharsEl.innerText = `${totalChars.toLocaleString()} chars`;
      if (totalTokensEl) totalTokensEl.innerText = `≈ ${totalTokens.toLocaleString()} tokens`;

      const pct = Math.min(100, Math.round((totalTokens / 1500) * 100));
      if (barFill) {
        barFill.style.width = `${pct}%`;
        if (totalTokens <= 1500) {
          barFill.style.background = 'linear-gradient(90deg, #10b981, #00f2fe)';
        } else {
          barFill.style.background = 'linear-gradient(90deg, #f59e0b, #f43f5e)';
        }
      }

      if (slaStatusEl) {
        if (totalTokens <= 1500) {
          slaStatusEl.className = 'sla-badge sla-pass';
          slaStatusEl.innerHTML = `<span>✓</span> SLA COMPLIANT (≤ 1,500 Tokens)`;
        } else {
          slaStatusEl.className = 'sla-badge sla-breach';
          slaStatusEl.innerHTML = `<span>⚠️</span> SLA BREACH (+${(totalTokens - 1500).toLocaleString()} tokens over budget)`;
        }
      }
    }

    [rootSlider, neighborSlider, callsSlider].forEach(slider => {
      slider.addEventListener('input', () => {
        recalculate();
      });
    });

    recalculate();
  }

  // ---------------------------------------------------------------------------
  // 17. REST API SANDBOX & CURL GENERATOR
  // ---------------------------------------------------------------------------
  function initApiPlayground() {
    const endpointSelect = document.getElementById('api-endpoint');
    const wsInput = document.getElementById('api-ws');
    const symInput = document.getElementById('api-sym');
    const qInput = document.getElementById('api-q');
    const curlBox = document.getElementById('api-curl-output');

    if (!endpointSelect || !curlBox) return;

    function updateCurl() {
      const ep = endpointSelect.value;
      const ws = wsInput ? wsInput.value.trim() : 'PolyTuring';
      const sym = symInput ? symInput.value.trim() : 'ContextCondenser';
      const q = qInput ? qInput.value.trim() : 'How does token budgeting work?';

      let cmd = '';

      switch (ep) {
        case 'query':
          cmd = `curl -s -X POST http://localhost:8080/api/query \\\n  -H "Content-Type: application/json" \\\n  -d '{"prompt": "${q}", "workspace": "${ws}"}'`;
          break;
        case 'condense':
          cmd = `curl -s "http://localhost:8080/api/condense?symbol=${sym}&workspace=${ws}&budget=1500"`;
          break;
        case 'symbol':
          cmd = `curl -s "http://localhost:8080/api/symbol?name=${sym}&workspace=${ws}"`;
          break;
        case 'references':
          cmd = `curl -s "http://localhost:8080/api/references?symbol=${sym}&workspace=${ws}"`;
          break;
        case 'galaxies':
          cmd = `curl -s "http://localhost:8080/api/galaxies?workspace=${ws}"`;
          break;
        case 'workspaces':
          cmd = `curl -s http://localhost:8080/api/workspaces`;
          break;
        case 'ingest':
          cmd = `curl -s -X POST http://localhost:8080/api/ingest \\\n  -H "Content-Type: application/json" \\\n  -d '{"path": "/path/to/${ws}", "refresh": true}'`;
          break;
      }

      curlBox.innerText = cmd;
    }

    endpointSelect.addEventListener('change', () => {
      updateCurl();
      HapticAudio.tap();
    });
    if (wsInput) wsInput.addEventListener('input', updateCurl);
    if (symInput) symInput.addEventListener('input', updateCurl);
    if (qInput) qInput.addEventListener('input', updateCurl);

    updateCurl();
  }

  // ---------------------------------------------------------------------------
  // 18. MERMAID DIAGRAMS RENDERING & SAFARI WEBKIT SIZING NORMALIZATION
  // ---------------------------------------------------------------------------
  function initMermaid() {
    if (window.mermaid) {
      mermaid.initialize({
        startOnLoad: true,
        theme: 'base',
        themeVariables: {
          darkMode: true,
          background: '#090d16',
          primaryColor: '#121a2f',
          primaryTextColor: '#f1f5f9',
          primaryBorderColor: '#00f2fe',
          lineColor: '#38bdf8',
          secondaryColor: '#17223b',
          secondaryTextColor: '#f1f5f9',
          secondaryBorderColor: '#a855f7',
          tertiaryColor: '#0c1220',
          tertiaryTextColor: '#f1f5f9',
          tertiaryBorderColor: '#334155',
          mainBkg: '#121a2f',
          nodeBorder: '#00f2fe',
          clusterBkg: 'rgba(15, 23, 42, 0.65)',
          clusterBorder: 'rgba(255, 255, 255, 0.15)',
          titleColor: '#00f2fe',
          edgeLabelBackground: '#0b0f19',
          fontSize: '13px',
          fontFamily: "'JetBrains Mono', 'Inter', monospace"
        }
      });

      // Normalize rendered SVGs to prevent Safari WebKit collapse
      setTimeout(() => {
        document.querySelectorAll('.mermaid svg').forEach(svg => {
          if (svg.viewBox && svg.viewBox.baseVal && svg.viewBox.baseVal.width > 0) {
            const vbW = svg.viewBox.baseVal.width;
            const vbH = svg.viewBox.baseVal.height;
            svg.setAttribute('width', vbW);
            svg.setAttribute('height', vbH);
            svg.style.aspectRatio = `${vbW} / ${vbH}`;
          }
        });
      }, 500);
    }
  }

})();
