import React from 'react';
import { Box, Container, CssBaseline } from '@mui/material';
import NavBar from './NavBar';

interface PageLayoutProps {
  children: React.ReactNode;
  maxWidth?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
}

export default function PageLayout({ children, maxWidth = 'lg' }: PageLayoutProps) {
  return (
    <>
      <CssBaseline />
      <NavBar />
      <Box component="main" sx={{ bgcolor: 'background.default', minHeight: '100vh' }}>
        <Container maxWidth={maxWidth} sx={{ pt: 4, pb: 6 }}>
          {children}
        </Container>
      </Box>
    </>
  );
}
