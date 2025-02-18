// Copyright 2024 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.
import React, { useEffect, useRef, useState } from 'react';
import { useDispatch } from 'react-redux';

import {
  Close as CloseIcon,
  SentimentVeryDissatisfied as DissatisfiedIcon,
  SentimentNeutral as NeutralIcon,
  SentimentSatisfiedAlt as SatisfiedIcon,
  SentimentVeryDissatisfiedOutlined as VeryDissatisfiedIcon,
  SentimentVerySatisfiedOutlined as VerySatisfiedIcon
} from '@mui/icons-material';
import {
  Button,
  Dialog,
  DialogContent,
  DialogTitle,
  IconButton,
  TextField,
  darken,
  dialogClasses,
  dialogTitleClasses,
  iconButtonClasses,
  lighten,
  textFieldClasses
} from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import actions from '@northern.tech/store/actions';
import { TIMEOUTS } from '@northern.tech/store/constants';
import { submitFeedback } from '@northern.tech/store/thunks';
import { isDarkMode } from '@northern.tech/store/utils';

const { setShowFeedbackDialog } = actions;

const useStyles = makeStyles()(theme => ({
  root: {
    pointerEvents: 'none',
    [`.${dialogClasses.paper}`]: { width: 350, bottom: 0, right: 0, position: 'absolute' },
    [`.${dialogTitleClasses.root}`]: {
      alignSelf: 'flex-end',
      padding: 0,
      [`.${iconButtonClasses.root}`]: { marginBottom: theme.spacing(-1) }
    },
    '.title': {
      color: isDarkMode(theme.palette.mode) ? lighten(theme.palette.primary.main, 0.85) : 'inherit'
    }
  },
  columns: { gap: theme.spacing(2) },
  rating: {
    [`.${iconButtonClasses.root}`]: {
      borderRadius: theme.shape.borderRadius,
      height: theme.spacing(6),
      width: theme.spacing(6),
      backgroundColor: isDarkMode(theme.palette.mode) ? darken(theme.palette.primary.main, 0.45) : lighten(theme.palette.primary.main, 0.85),
      color: theme.palette.primary.main,
      '&:hover': {
        backgroundColor: theme.palette.primary.main,
        color: lighten(theme.palette.primary.main, 0.85)
      }
    }
  },
  text: { [`.${textFieldClasses.root}`]: { marginTop: 0 }, '.submitButton': { alignSelf: 'start' } }
}));

const satisfactionLevels = [
  { Icon: VeryDissatisfiedIcon, title: 'Very Dissatisfied' },
  { Icon: DissatisfiedIcon, title: 'Dissatisfied' },
  { Icon: NeutralIcon, title: 'Neutral' },
  { Icon: SatisfiedIcon, title: 'Satisfied' },
  { Icon: VerySatisfiedIcon, title: 'Very Satisfied' }
];
const explanations = ['Very unsatisfied', 'Very satisfied'];

const SatisfactionGauge = ({ classes, setSatisfaction }) => (
  <div className={`flexbox column ${classes.columns}`}>
    <div className="title">How satisfied are you with Mender?</div>
    <div className={`flexbox space-between ${classes.rating}`}>
      {satisfactionLevels.map(({ Icon, title }, index) => (
        <IconButton key={`satisfaction-${index}`} onClick={() => setSatisfaction(index)} title={title}>
          <Icon fontSize="large" />
        </IconButton>
      ))}
    </div>
    <div className="flexbox space-between muted">
      {explanations.map((explanation, index) => (
        <div className="slightly-smaller" key={`explanation-${index}`}>
          {explanation}
        </div>
      ))}
    </div>
  </div>
);

const TextEntry = ({ classes, feedback, onChangeFeedback, onSubmit }) => (
  <div className={`flexbox column ${classes.columns} ${classes.text}`}>
    <div className="title">What do you think is the most important thing to improve in Mender? (optional)</div>
    <TextField
      placeholder="Your feedback"
      multiline
      minRows={4}
      onChange={({ target: { value } }) => onChangeFeedback(value)}
      value={feedback}
      variant="outlined"
    />
    <Button className="submitButton" variant="contained" onClick={onSubmit}>
      Submit Feedback
    </Button>
  </div>
);

const AppreciationNote = () => <p className="margin-top-none align-center title">Thank you for taking the time to share your thoughts!</p>;

const progressionLevels = [SatisfactionGauge, TextEntry, AppreciationNote];

export const FeedbackDialog = () => {
  const [progress, setProgress] = useState(0);
  const [satisfaction, setSatisfaction] = useState(-1);
  const [feedback, setFeedback] = useState('');
  const dispatch = useDispatch();
  const isInitialized = useRef(false);

  const { classes } = useStyles();

  useEffect(() => {
    if (!isInitialized.current) {
      return;
    }
    setProgress(current => current + 1);
  }, [satisfaction]);

  useEffect(() => {
    setTimeout(() => (isInitialized.current = true), TIMEOUTS.oneSecond);
  }, []);

  const onCloseClick = () => dispatch(setShowFeedbackDialog(false));

  const onSubmit = () => {
    setProgress(progress + 1);
    dispatch(submitFeedback({ satisfaction: satisfactionLevels[satisfaction].title, feedback }));
  };

  const Component = progressionLevels[progress];
  return (
    <Dialog className={classes.root} open hideBackdrop disableEnforceFocus PaperProps={{ style: { pointerEvents: 'auto' } }}>
      <DialogTitle>
        <IconButton onClick={onCloseClick} aria-label="close" size="small">
          <CloseIcon />
        </IconButton>
      </DialogTitle>
      <DialogContent>
        <Component classes={classes} feedback={feedback} setSatisfaction={setSatisfaction} onChangeFeedback={setFeedback} onSubmit={onSubmit} />
      </DialogContent>
    </Dialog>
  );
};

export default FeedbackDialog;
