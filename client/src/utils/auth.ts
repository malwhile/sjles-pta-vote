import axios from 'axios';

/**
 * Call the logout endpoint and clear local authentication state
 */
export async function logout(): Promise<void> {
  const token = localStorage.getItem('adminToken');

  try {
    if (token) {
      // Call logout endpoint for audit logging
      await axios.post(
        '/api/admin/logout',
        {},
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }
      );
    }
  } catch (error) {
    // Log error but continue with local logout
    console.error('Logout API call failed:', error);
  } finally {
    // Always clear local state
    localStorage.removeItem('adminToken');
  }
}
