import { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Link } from 'react-router-dom';

import { Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { getOrganization } from '@northern.tech/store/selectors';
import { setOnboardingApproach } from '@northern.tech/store/thunks';

import CopyCode from '../../../common-ui/CopyCode';

const useStyles = makeStyles()(() => ({
  link: {
    textDecoration: 'underline'
  }
}));
export const McuDeviceOnboarding = () => {
  const { tenant_token: tenantToken } = useSelector(getOrganization);
  const dispatch = useDispatch();
  const { classes } = useStyles();
  useEffect(() => {
    dispatch(setOnboardingApproach('mcu'));
  }, [dispatch]);

  return (
    <div>
      <Typography variant="subtitle1" className="margin-bottom-small">
        Set up a Zephyr application with Mender
      </Typography>
      <Typography className="margin-bottom-small">
        {' '}
        To do this, you’ll need to visit our documentation and follow the guide there. Go to{' '}
        <a
          className={classes.link}
          href="https://docs.mender.io/get-started/microcontroller-preview/prepare-an-esp32-s3-with-zephyr"
          target="_blank"
          rel="noopener noreferrer"
        >
          Prepare an ESP32-S3 with Zephyr
        </a>
      </Typography>
      <Typography className="margin-bottom-small" variant="body1">
        First, copy your <b>organization token</b> below as you will need it to complete the steps in the guide:
      </Typography>
      <CopyCode code={tenantToken} withDescription={true} />
      <Typography variant="body1">
        Once you have completed steps 1 and 2 in the documentation, your device should show in the Mender UI as <i>“Pending”</i>. Go to{' '}
        <Link className={classes.link} to="/devices">
          Devices
        </Link>{' '}
        to see it.
      </Typography>
    </div>
  );
};
