import { createTheme } from '@mui/material/styles';

export const theme = createTheme({
  palette: {
    primary: {
      main: '#2E7FD9',
    },
    secondary: {
      main: '#D4AF37',
    },
    background: {
      default: '#F5F7FA',
      paper: '#FFFFFF',
    },
    success: {
      main: '#28a745',
    },
    error: {
      main: '#dc3545',
    },
  },
  components: {
    MuiAppBar: {
      styleOverrides: {
        root: {
          boxShadow: 'none',
          borderBottom: '1px solid #e0e0e0',
        },
      },
    },
    MuiButton: {
      styleOverrides: {
        root: {
          textTransform: 'none',
          boxShadow: 'none',
          borderRadius: '6px',
          '&:hover': {
            boxShadow: 'none',
          },
        },
      },
    },
    MuiCard: {
      styleOverrides: {
        root: {
          borderRadius: '12px',
          boxShadow: '0 1px 3px rgba(0, 0, 0, 0.1)',
        },
      },
    },
    MuiTextField: {
      defaultProps: {
        variant: 'outlined',
        size: 'small',
      },
    },
    MuiTableHead: {
      styleOverrides: {
        root: {
          backgroundColor: '#2E7FD9',
          '& th': {
            color: 'white',
            fontWeight: 600,
          },
        },
      },
    },
  },
});
