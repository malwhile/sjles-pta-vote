import React, { useState } from 'react';
import {
  AppBar,
  Toolbar,
  Typography,
  IconButton,
  Drawer,
  List,
  ListItemButton,
  Box,
  Divider,
} from '@mui/material';
import MenuIcon from '@mui/icons-material/Menu';
import { Link as RouterLink } from 'react-router';

export default function NavBar() {
  const [drawerOpen, setDrawerOpen] = useState(false);

  const handleDrawerToggle = () => {
    setDrawerOpen(!drawerOpen);
  };

  const closeDrawer = () => {
    setDrawerOpen(false);
  };

  const navigationItems = [
    { label: 'Home', path: '/' },
    { label: 'Admin Login', path: '/admin-login' },
  ];

  const memberItems = [
    { label: 'Upload Members', path: '/admin-members' },
    { label: 'View Members', path: '/admin-members-view' },
  ];

  const voteItems = [
    { label: 'Create Vote', path: '/create-vote' },
    { label: 'Poll List', path: '/polls' },
  ];

  return (
    <>
      <AppBar position="static">
        <Toolbar>
          <IconButton
            color="inherit"
            edge="start"
            onClick={handleDrawerToggle}
            aria-label="menu"
          >
            <MenuIcon />
          </IconButton>
          <Box sx={{ flexGrow: 1, display: 'flex', justifyContent: 'center' }}>
            <Typography variant="h6" sx={{ fontWeight: 600 }}>
              🐾 SJLES PTA Voting 🐾
            </Typography>
          </Box>
          <Box sx={{ width: 48 }} />
        </Toolbar>
      </AppBar>

      <Drawer anchor="left" open={drawerOpen} onClose={closeDrawer}>
        <Box sx={{ width: 280, pt: 2 }}>
          <List>
            {navigationItems.map((item) => (
              <ListItemButton
                key={item.path}
                component={RouterLink}
                to={item.path}
                onClick={closeDrawer}
              >
                {item.label}
              </ListItemButton>
            ))}

            <Divider sx={{ my: 1 }} />

            <ListItemButton disabled sx={{ fontWeight: 600, color: 'textPrimary', cursor: 'default' }}>
              Member
            </ListItemButton>
            {memberItems.map((item) => (
              <ListItemButton
                key={item.path}
                component={RouterLink}
                to={item.path}
                onClick={closeDrawer}
                sx={{ pl: 4 }}
              >
                {item.label}
              </ListItemButton>
            ))}

            <Divider sx={{ my: 1 }} />

            <ListItemButton disabled sx={{ fontWeight: 600, color: 'textPrimary', cursor: 'default' }}>
              Vote
            </ListItemButton>
            {voteItems.map((item) => (
              <ListItemButton
                key={item.path}
                component={RouterLink}
                to={item.path}
                onClick={closeDrawer}
                sx={{ pl: 4 }}
              >
                {item.label}
              </ListItemButton>
            ))}
          </List>
        </Box>
      </Drawer>
    </>
  );
}
