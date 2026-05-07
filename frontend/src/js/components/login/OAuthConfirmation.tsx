import { useState } from 'react';
import { Link as RouterLink } from 'react-router';

import { Alert, Button, Paper, TextField, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import DotsGradient from '../../../assets/img/dots-gradient.svg';
import LoginLogo from '../../../assets/img/loginlogo.svg';
import DotsWhite from '../../../assets/img/verymuch.svg';

const useStyles = makeStyles()(theme => ({
  loginBox: { maxWidth: 525, justifySelf: 'center' },
  headerWrapper: {
    display: 'grid',
    gridTemplateColumns: '1fr auto 1fr',
    alignItems: 'center'
  },
  header: {
    fontSize: '2.125rem',
    fontWeight: 500
  },
  ntBrandingLeft: {
    top: 0,
    left: 0,
    bottom: 0,
    width: '25%',
    maxWidth: 365,
    overflow: 'hidden',
    zIndex: -2,
    '&::before': {
      content: '""',
      position: 'absolute',
      inset: 0,
      background: '#284D68',
      borderTopRightRadius: 120,
      transform: 'skewY(3deg)',
      transformOrigin: '0 0'
    }
  },
  ntBrandingRight: {
    right: 0,
    bottom: 0,
    width: 88,
    height: 289,
    overflow: 'hidden',
    zIndex: -2,
    '&::before': {
      content: '""',
      display: 'block',
      width: 638,
      height: 704,
      background: 'linear-gradient(to top, #970F57, #02AFCF)',
      borderTopLeftRadius: 120
    }
  },
  dotsOverlayTop: {
    top: -10,
    right: -60,
    zIndex: 1,
    pointerEvents: 'none'
  },
  dotsOverlayBottom: {
    bottom: -24,
    left: -40,
    zIndex: 1,
    pointerEvents: 'none',
    '& path': { stroke: theme.palette.primary.light }
  },
  wideAlert: {
    width: 600,
    maxWidth: 'none'
  }
}));

type Step = 'confirm' | 'verify';

type OAuthConfirmationProps = {
  email?: string;
  provider?: string;
};

type ConfirmStepProps = {
  classes: { header: string };
  onContinue: () => void;
  provider: string;
};

const ConfirmStep = ({ classes, provider, onContinue }: ConfirmStepProps) => (
  <>
    <Typography className={classes.header}>Link {provider} account?</Typography>
    <Typography className="margin-top-small">
      By linking your {provider} account, we will switch your Mender account&#39;s authentication method from email and password to {provider} OAuth. Moving
      forward, your credentials, personal information, and two-factor authentication will be managed through your {provider} account.
    </Typography>
    <div className="margin-top-medium flexbox">
      <Button variant="outlined" color="secondary" component={RouterLink} to="/login">
        Cancel
      </Button>
      <Button className="margin-left-small" variant="contained" color="secondary" onClick={onContinue}>
        Link my accounts
      </Button>
    </div>
  </>
);

type VerifyStepProps = {
  classes: { header: string; wideAlert: string };
  email: string;
  onCancel: () => void;
  provider: string;
};

const VerifyStep = ({ classes, provider, email, onCancel }: VerifyStepProps) => {
  const [password, setPassword] = useState('');
  return (
    <>
      <Typography className={classes.header}>Verify your identity</Typography>
      <Alert severity="warning" className={`margin-top-small ${classes.wideAlert}`}>
        Verify your identity by entering your password for your Mender account below.
      </Alert>
      <Typography className="margin-top-small">This is the {provider} email address we matched to your Mender account.</Typography>
      <TextField label="Your email" value={email} disabled fullWidth margin="normal" />
      <TextField label="Password" type="password" value={password} onChange={event => setPassword(event.target.value)} fullWidth margin="normal" />
      <div className="margin-top-medium flexbox">
        <Button variant="outlined" color="secondary" onClick={onCancel}>
          Cancel
        </Button>
        <Button className="margin-left-small" variant="contained" color="secondary">
          Confirm
        </Button>
      </div>
    </>
  );
};

export const OAuthConfirmation = ({ provider = 'Microsoft', email = 'chiacheng.lu@northern.tech' }: OAuthConfirmationProps) => {
  const [step, setStep] = useState<Step>('confirm');
  const { classes } = useStyles();
  return (
    <>
      <div className={classes.headerWrapper}>
        <div className="margin-top-large margin-left-large" />
      </div>
      <Paper elevation={0} className={`flexbox margin-top-large padding-small column ${classes.loginBox}`}>
        <LoginLogo className="margin-bottom-large" style={{ width: 210 }} />
        {step === 'confirm' ? (
          <ConfirmStep classes={classes} provider={provider} onContinue={() => setStep('verify')} />
        ) : (
          <VerifyStep classes={classes} provider={provider} email={email} onCancel={() => setStep('confirm')} />
        )}
      </Paper>
      <div className={`absolute ${classes.ntBrandingLeft}`}>
        <DotsGradient className={`absolute ${classes.dotsOverlayTop}`} />
        <DotsWhite className={`absolute ${classes.dotsOverlayBottom}`} />
      </div>
      <div className={`absolute ${classes.ntBrandingRight}`} />
    </>
  );
};
