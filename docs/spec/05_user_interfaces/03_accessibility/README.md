# Accessibility

AI Together interfaces are designed to be accessible to users with disabilities.

## Purpose

Ensure all users can effectively use AI Together, regardless of ability.

## Status

**⚠️ Deferred** - This feature is planned for future implementation. Current priority is on core functionality.

## Accessibility Standards

AI Together aims to meet:
- **WCAG 2.1 Level AA** - Web Content Accessibility Guidelines
- **Section 508** - U.S. federal accessibility requirements

## Member Desktop App

### Keyboard Navigation

**Full Keyboard Access:**
- All features accessible via keyboard
- Tab order follows logical flow
- Keyboard shortcuts for common actions
- Visible focus indicator on all interactive elements

**Keyboard Shortcuts:**
| Shortcut | Action |
|----------|--------|
| `Tab` | Move between sections |
| `Enter/Space` | Activate focused item |
| `Escape` | Close modal/dialog |
| `Cmd/Ctrl+,` | Open settings |
| `Cmd/Ctrl+L` | Open logs |

### Screen Reader Support

**Compatibility:**
- Windows: Narrator, NVDA, JAWS
- macOS: VoiceOver
- Linux: Orca

**Screen readers can announce:**
- All text content
- Button and link labels
- Status indicators (online/offline)
- Form fields and their purpose
- Error messages

### Visual Accessibility

**Color Contrast:**
- All text meets WCAG AA contrast ratios (4.5:1 for normal text)
- Charts use distinct colors with patterns/labels
- Status indicators use shapes + colors (not color alone)

**Text Sizing:**
- UI supports system text scaling
- Minimum font size: 12px
- Text is readable at 200% zoom

### High Contrast Mode

**Available on:**
- Windows: High contrast mode
- macOS: Increase contrast
- Linux: High contrast theme

**Behavior:**
- Interface adapts to system settings
- Essential information remains visible

## Manager Dashboard

### Keyboard Navigation

**Full Keyboard Access:**
- Tab through all interactive elements
- Skip links for main navigation
- Visible focus indicators
- No keyboard traps

### Screen Reader Support

**Semantic HTML:**
- Proper heading structure (h1, h2, h3...)
- Labels for all form inputs
- ARIA labels for interactive elements
- ARIA live regions for dynamic updates

### Color Independence

**Information Not Conveyed by Color Alone:**
- Charts include data labels
- Status indicators use icons + colors
- Error messages use text + styling
- Tables use borders/striping, not just color

### Focus Management

**When Modal Opens:**
- Focus moves to modal
- Background content is hidden from screen readers
- Tab cycles within modal only
- Escape closes modal and returns focus

### Form Accessibility

**All Forms Include:**
- Required field indicators
- Clear labels
- Validation error messages
- Instructions for complex inputs
- Sufficient time to complete (no timeouts)

## Functional Requirements

- **FR-001:** All features must be keyboard accessible
- **FR-002:** Focus must be visible on all interactive elements
- **FR-003:** Screen readers must announce all important information
- **FR-004:** Color alone must not convey information
- **FR-005:** Text contrast must meet WCAG AA standards

## Testing

**Regular Accessibility Testing:**
- Keyboard-only navigation
- Screen reader testing (NVDA, VoiceOver)
- Color contrast verification
- Zoom testing (up to 200%)

## Known Limitations

**Current Limitations:**
- Charts may not have full screen reader descriptions (data tables provided as alternative)
- Some interactive elements may not have full ARIA descriptions

**Planned Improvements:**
- Enhanced chart accessibility
- Full ARIA label coverage
- Comprehensive accessibility testing

## Reporting Accessibility Issues

**If you encounter accessibility barriers:**
- Report via: support@code-together.com
- Include: Description of issue, assistive technology used, steps to reproduce
- We will respond within 5 business days

## Success Criteria

- Keyboard-only users can complete all tasks
- Screen reader users can access all information
- Users with color blindness can distinguish all information
- Low-vision users can use zoom up to 200%

---

**Related:** [05.01 Member App](../01_member_app/) | [05.02 Manager Dashboard](../02_manager_dashboard/)
