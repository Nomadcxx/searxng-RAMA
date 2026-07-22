## Multi-Theme Design Plan for SearXNG RAMA

### 1. Vision & Objectives

**Core Purpose**: Create a distinctive search interface that feels Google-polished while maintaining RAMA's privacy-first ethos and minimalist aesthetic.

**User Experience Goals**:
- Professional, intuitive interface that feels familiar (Google-inspired)
- Multiple theme options for user preference
- Seamless switching between themes
- Consistent visual language across all themes
- Maintained performance and accessibility standards

### 2. Theme Architecture Design

#### Theme Structure
```
themes/
├── rama/              # Default dark theme (current)
│   ├── definitions.less    # Core variables
│   ├── dark.less          # Dark variant
│   └── components.less    # UI components
├── google/             # New Google-inspired theme
│   ├── definitions.less    # Core variables
│   ├── light.less         # Light variant
│   ├── dark.less          # Dark variant (Google-style)
│   └── components.less    # UI components
└── shared/             # Common elements
    └── base-components.less # Shared UI patterns
```

### 3. Google Theme Specifications

#### Aesthetic Direction: "Refined Minimalism"
- Clean, professional appearance inspired by Google's 2024 UI
- Focus on typography, spacing, and subtle visual hierarchy
- No skeuomorphic elements or excessive decoration
- Emphasis on functionality with elegant presentation

#### Color System
**Light Variant**:
- Background: #FFFFFF (pure white)
- Text: #202124 (Google's standard dark gray)
- Links: #1a0dab (Google blue)
- Visited: #681da8 (Google purple)
- Accent: #ef233c (RAMA red for CTAs/accents)
- Borders: #dadce0 (subtle gray)
- Cards: #ffffff with subtle shadows

**Dark Variant**:
- Background: #202124 (Google's dark background)
- Text: #e8eaed (light gray)
- Links: #8ab4f8 (Google's dark blue)
- Visited: #c58af9 (Google purple)
- Accent: #ef233c (RAMA red for CTAs/accents)
- Borders: #5f6368 (medium gray)

#### Typography System
**Font Stack**: 
- Display: System fonts stack (preserving privacy)
- Body: System UI fonts with Roboto-inspired fallbacks
- Mono: JetBrains Mono (maintaining current RAMA choice)

**Type Scale**:
- XS: 0.75rem (12px) - Metadata, footnotes
- SM: 0.875rem (14px) - Body text, labels
- Base: 1rem (16px) - Main content
- MD: 1.125rem (18px) - Subheadings
- LG: 1.25rem (20px) - Headings
- XL: 1.5rem (24px) - Section titles
- 2XL: 1.875rem (30px) - Page titles
- 3XL: 2.25rem (36px) - Hero text

#### Spacing System (8px Grid)
- 0: 0rem
- 1: 0.25rem (4px)
- 2: 0.5rem (8px)
- 3: 0.75rem (12px)
- 4: 1rem (16px)
- 5: 1.25rem (20px)
- 6: 1.5rem (24px)
- 8: 2rem (32px)
- 10: 2.5rem (40px)
- 12: 3rem (48px)

### 4. Component Design Specifications

#### Search Bar
- Google-inspired rounded corners (8px radius)
- Subtle white background with light gray border (#dadce0)
- On focus: 2px red border (#ef233c), subtle elevation
- Clear button fades in when text present
- Autocomplete dropdown with smooth animation

#### Result Cards
- Clean card layout with 8px rounded corners
- Subtle border (#e8eaed) or none for clean look
- Hover effect: 2px upward movement + shadow
- Clear visual hierarchy: URL > Title > Snippet
- Favicon integration with proper sizing
- Structured data display with consistent spacing

#### Buttons & Controls
- Google-style pill-shaped buttons for actions
- Rounded rectangles for secondary actions
- Subtle hover effects (slight lift, shadow intensification)
- Clear active states with pressed appearance
- Icon buttons with consistent sizing

#### Typography & Visual Hierarchy
- Strong URL differentiation with blue color
- Published dates and metadata in lighter gray
- Clear separation between results with whitespace
- Consistent vertical rhythm throughout interface

### 5. Implementation Strategy

#### Phase 1: Core Infrastructure
1. Create theme directory structure
2. Implement theme selection mechanism in PKGBUILD
3. Update SearXNG configuration for multi-theme support
4. Create shared component base

#### Phase 2: Google Theme Development
1. Implement Google theme variables (light/dark)
2. Design UI components with new aesthetic
3. Add micro-interactions and animations
4. Ensure accessibility compliance

#### Phase 3: Installer Integration
1. Add theme selection to TUI installer
2. Implement theme compilation in build process
3. Test theme switching functionality
4. Document usage for end users

#### Phase 4: Testing & Refinement
1. Cross-browser compatibility testing
2. Performance impact assessment
3. Accessibility audit (contrast, keyboard nav)
4. User feedback collection and iteration

### 6. Technical Considerations

#### Privacy Preservation
- Continue using system fonts only
- No external CDN requests
- No tracking or analytics
- Local compilation only

#### Performance Requirements
- CSS bundle size under 50KB per theme
- No JavaScript dependencies for core functionality
- Efficient CSS selectors (avoid nesting > 3 levels)
- Hardware-accelerated animations where possible

#### Accessibility Standards
- WCAG 2.1 AA compliance (minimum 4.5:1 contrast)
- Keyboard navigation support
- Screen reader compatibility
- Reduced motion preferences respected

#### Maintainability
- Consistent naming conventions (BEM methodology)
- Clear variable organization (by category/type)
- Comprehensive comments for future developers
- Version-controlled theme development

### 7. Success Metrics

**Quantitative**:
- Theme switching works seamlessly without reload
- CSS file sizes under 50KB each
- Page load time increase < 100ms
- Contrast ratios meet WCAG AA standards

**Qualitative**:
- Google-polished aesthetic achieved
- RAMA identity preserved through accent colors
- Users can easily distinguish between theme options
- Interface feels professional yet minimalist
- Micro-interactions enhance rather than distract

### 8. Next Steps

With this design plan established, we can now proceed to:
1. Implement the theme directory structure
2. Create the Google-inspired theme files
3. Integrate theme selection into the installer
4. Test and refine the implementation

This approach ensures we have a solid foundation for multiple themes while maintaining the quality and privacy focus that RAMA is known for.