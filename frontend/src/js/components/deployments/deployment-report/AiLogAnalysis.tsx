// Copyright 2025 Northern.tech AS
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
import { useRef, useState } from 'react';
import { useSelector } from 'react-redux';
import { Link } from 'react-router-dom';

import {
  AutoAwesomeOutlined as AutoAwesomeIcon,
  ContentCopyOutlined as CopyPasteIcon,
  ThumbDownOutlined as ThumbDownIcon,
  ThumbUpOutlined as ThumbUpIcon
} from '@mui/icons-material';
import { Alert, Button, IconButton, Slide, Typography, alpha } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { SparkleAnimation } from '@northern.tech/common-ui/Sparkles';
import { getGlobalSettings, getUserRoles } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { generateDeploymentLogAnalysis, submitUserFeedback } from '@northern.tech/store/thunks';
import copy from 'copy-to-clipboard';
import dayjs from 'dayjs';
import duration from 'dayjs/plugin/duration.js';
import relativeTime from 'dayjs/plugin/relativeTime.js';
import { MarkdownToJSX } from 'markdown-to-jsx';
import MuiMarkdown, { defaultOverrides } from 'mui-markdown';

dayjs.extend(duration);
dayjs.extend(relativeTime);

const useStyles = makeStyles()(theme => ({
  alert: { display: 'inline-flex', marginBottom: theme.spacing(2) },
  analysisResult: {
    backgroundColor: alpha(theme.palette.secondary.light, 0.08),
    border: `1px solid ${theme.palette.secondary.light}`,
    borderRadius: theme.shape.borderRadius,
    padding: theme.spacing(2)
  }
}));

const markDownStyleOverrides: MarkdownToJSX.Overrides = {
  ...defaultOverrides,
  h1: { component: 'b' },
  h2: { component: 'b' },
  h3: { component: 'b' }
};

interface AiLogAnalysisProps {
  deployment: { devices?: Record<string, unknown>; id: string };
  deviceId: string;
}

const Header = () => (
  <div className="flexbox center-aligned margin-bottom-small">
    <AutoAwesomeIcon color="secondary" className="margin-right-small" />
    <Typography variant="h6">AI summary (experimental)</Typography>
  </div>
);

const FeedbackSection = ({ deploymentId, deviceId }) => {
  const [feedbackSubmitted, setFeedbackSubmitted] = useState(false);
  const dispatch = useAppDispatch();
  const handleFeedback = isHelpful => {
    dispatch(submitUserFeedback({ formId: 'feat.ai', feedback: { useful: isHelpful, deployment_id: deploymentId, device_id: deviceId }}))
    setFeedbackSubmitted(true);
  };

  return (
    <div className="flexbox center-aligned">
      {feedbackSubmitted ? (
        <Typography variant="body2" color="text.secondary">
          Thank you for your feedback!
        </Typography>
      ) : (
        <>
          <Typography className="margin-right-small" variant="body2" color="text.secondary">
            Was this helpful?
          </Typography>
          <IconButton size="small" aria-label="thumbs-up" onClick={() => handleFeedback(true)}>
            <ThumbUpIcon fontSize="small" />
          </IconButton>
          <IconButton size="small" aria-label="thumbs-down" onClick={() => handleFeedback(false)}>
            <ThumbDownIcon fontSize="small" />
          </IconButton>
        </>
      )}
    </div>
  );
};

const AiNotEnabledNote = ({ className, isAdmin }) => (
  <Alert className={className} severity="info">
    AI features are not enabled for this organization.
    {isAdmin ? (
      <>
        Go to the settings page to enable this feature.
        <Link className="margin-left-small" to="/settings/global-settings">
          Settings
        </Link>
      </>
    ) : (
      'Contact your admin to enable this feature.'
    )}
  </Alert>
);

export const AiLogAnalysis = ({ deployment, deviceId }: AiLogAnalysisProps) => {
  const { classes } = useStyles();
  const dispatch = useAppDispatch();
  const [isAnalyzing, setIsAnalyzing] = useState(false);
  const [analysisResult, setAnalysisResult] = useState('');
  const [analysisError, setAnalysisError] = useState<string | null>(null);
  const { aiFeatures = {} } = useSelector(getGlobalSettings);
  const { isAdmin } = useSelector(getUserRoles);

  const { enabled: isAiEnabled } = aiFeatures;
  const slideOutRef = useRef<HTMLDivElement | null>(null);

  const onGenerateAnalysisClick = async () => {
    setIsAnalyzing(true);
    setAnalysisError(null);
    try {
      const result = await dispatch(generateDeploymentLogAnalysis({ deploymentId: deployment.id, deviceId })).unwrap();
      setAnalysisResult(result);
      setIsAnalyzing(false);
    } catch (error) {
      console.error('Error generating analysis:', error);
      setIsAnalyzing(false);
      if (error.status === 429) {
        const waitingTime = dayjs.duration(error.request.getResponseHeader('Retry-After'), 'seconds').humanize();
        setAnalysisError(`You have reached your limit of 50 AI requests per day. Please try again in about ${waitingTime}.`);
        return;
      }
      setAnalysisError('Failed to generate analysis. Please try again.');
    }
  };

  const onCopyAnalysisClick = () => copy(analysisResult);

  return (
    <div className="padding-top-small padding-bottom-small">
      <Header />
      <div className="flexbox center-aligned">
        <Button className="margin-right-small" color="secondary" disabled={!isAiEnabled || isAnalyzing} onClick={onGenerateAnalysisClick} variant="contained">
          {isAnalyzing ? 'Generating summary...' : 'Generate summary'}
        </Button>
      </div>
      <div className="margin-top-small">
        {!isAiEnabled && <AiNotEnabledNote className={classes.alert} isAdmin={isAdmin} />}
        {analysisError && (
          <Alert className={classes.alert} severity="error">
            {analysisError}
          </Alert>
        )}
      </div>
      <div ref={slideOutRef}>
        <Slide in={!!analysisResult || isAnalyzing} container={slideOutRef.current}>
          <div className={`fadeInSlow ${classes.analysisResult}`}>
            {!analysisResult ? (
              <div className="flexbox">
                <SparkleAnimation className="margin-right-x-small" />
                <Typography variant="body1" color="text.secondary">
                  Thinking...
                </Typography>
              </div>
            ) : (
              <>
                <>
                  <div className="flexbox space-between center-aligned">
                    <div className="flexbox center-aligned">
                      <AutoAwesomeIcon className="margin-right-small" />
                      Summary of Deployment Failure:
                    </div>
                    <IconButton onClick={onCopyAnalysisClick}>
                      <CopyPasteIcon />
                    </IconButton>
                  </div>
                  <div className="margin">
                    <MuiMarkdown overrides={markDownStyleOverrides}>{analysisResult}</MuiMarkdown>
                  </div>
                </>
                <div className="flexbox center-aligned space-between">
                  <FeedbackSection deploymentId={deployment.id} deviceId={deviceId} />
                  <Typography variant="body2" color="text.secondary">
                    AI-generated information can be inaccurate, so always verify it before taking action.
                  </Typography>
                </div>
              </>
            )}
          </div>
        </Slide>
      </div>
    </div>
  );
};
