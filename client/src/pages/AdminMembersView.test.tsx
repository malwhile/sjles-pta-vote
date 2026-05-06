import { render, screen, waitFor } from '@testing-library/react';
import AdminMembersView from './AdminMembersView';

// Mock react-router
jest.mock('react-router', () => ({
  useNavigate: () => jest.fn(),
}));

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => {
      store[key] = value.toString();
    },
    removeItem: (key: string) => {
      delete store[key];
    },
    clear: () => {
      store = {};
    },
  };
})();

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
});

// Mock fetch
global.fetch = jest.fn();
const mockFetch = global.fetch as jest.MockedFunction<typeof fetch>;

const renderComponent = () => {
  return render(<AdminMembersView />);
};

describe('AdminMembersView', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    localStorage.clear();
    mockFetch.mockClear();
  });

  test('renders without crashing when authenticated', async () => {
    localStorage.setItem('adminToken', 'test-token');

    mockFetch.mockImplementation((url: string | RequestInfo) => {
      const urlStr = typeof url === 'string' ? url : url.toString();
      if (urlStr.includes('/api/admin/members/years')) {
        return Promise.resolve(
          new Response(
            JSON.stringify({ success: true, data: [2025, 2024, 2023] }),
            { status: 200 }
          )
        );
      }
      return Promise.reject(new Error('Not found'));
    });

    renderComponent();

    // Wait for component to render
    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalled();
    });
  });

  test('fetches years from API endpoint on mount', async () => {
    localStorage.setItem('adminToken', 'test-token');

    mockFetch.mockImplementation((url: string | RequestInfo) => {
      const urlStr = typeof url === 'string' ? url : url.toString();
      if (urlStr.includes('/api/admin/members/years')) {
        return Promise.resolve(
          new Response(
            JSON.stringify({ success: true, data: [2025, 2024, 2023] }),
            { status: 200 }
          )
        );
      }
      return Promise.reject(new Error('Not found'));
    });

    renderComponent();

    await waitFor(() => {
      const calls = (mockFetch as jest.Mock).mock.calls;
      const yearsCall = calls.some((call: any[]) =>
        typeof call[0] === 'string' && call[0].includes('/api/admin/members/years')
      );
      expect(yearsCall).toBe(true);
    });
  });

  test('includes authorization header in API request', async () => {
    const testToken = 'test-auth-token-12345';
    localStorage.setItem('adminToken', testToken);

    mockFetch.mockImplementation((url: string | RequestInfo) => {
      const urlStr = typeof url === 'string' ? url : url.toString();
      if (urlStr.includes('/api/admin/members/years')) {
        return Promise.resolve(
          new Response(
            JSON.stringify({ success: true, data: [2025] }),
            { status: 200 }
          )
        );
      }
      return Promise.reject(new Error('Not found'));
    });

    renderComponent();

    await waitFor(() => {
      const calls = (mockFetch as jest.Mock).mock.calls;
      const yearsCall = calls.find((call: any[]) =>
        typeof call[0] === 'string' && call[0].includes('/api/admin/members/years')
      );
      expect(yearsCall).toBeDefined();

      if (yearsCall && yearsCall[1]) {
        const headers = yearsCall[1].headers as any;
        expect(headers['Authorization']).toBe(`Bearer ${testToken}`);
      }
    });
  });

  test('handles empty years array from API', async () => {
    localStorage.setItem('adminToken', 'test-token');

    mockFetch.mockImplementation((url: string | RequestInfo) => {
      const urlStr = typeof url === 'string' ? url : url.toString();
      if (urlStr.includes('/api/admin/members/years')) {
        return Promise.resolve(
          new Response(
            JSON.stringify({ success: true, data: [] }),
            { status: 200 }
          )
        );
      }
      return Promise.reject(new Error('Not found'));
    });

    renderComponent();

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalled();
    });
  });

  test('handles API error gracefully', async () => {
    localStorage.setItem('adminToken', 'test-token');

    mockFetch.mockImplementation((url: string | RequestInfo) => {
      const urlStr = typeof url === 'string' ? url : url.toString();
      if (urlStr.includes('/api/admin/members/years')) {
        return Promise.reject(new Error('Network error'));
      }
      return Promise.reject(new Error('Not found'));
    });

    // Should not throw
    renderComponent();

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalled();
    });
  });

  test('handles 401 unauthorized response', async () => {
    localStorage.setItem('adminToken', 'expired-token');

    mockFetch.mockImplementation((url: string | RequestInfo) => {
      const urlStr = typeof url === 'string' ? url : url.toString();
      if (urlStr.includes('/api/admin/members/years')) {
        return Promise.resolve(
          new Response(
            JSON.stringify({ success: false, error: 'Unauthorized' }),
            { status: 401 }
          )
        );
      }
      return Promise.reject(new Error('Not found'));
    });

    renderComponent();

    await waitFor(() => {
      // Token should be cleared after 401
      expect(localStorage.getItem('adminToken')).toBeNull();
    });
  });
});
