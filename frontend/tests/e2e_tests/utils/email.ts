// Copyright 2026 Northern.tech AS
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
import { OAuth2Client } from 'google-auth-library';
import { gmail_v1, google } from 'googleapis';

interface Email {
  body: string;
  from: string;
  id: string;
  receivedAt: Date;
  subject: string;
  to: string[];
}

interface SearchParams {
  from?: string;
  to?: string;
  unread?: boolean; // filter only unread messages
}

export interface EmailClient {
  getEmails(params?: SearchParams): Promise<Email[]>;
}

export class GmailEmailClient implements EmailClient {
  private gmail: gmail_v1.Gmail;
  private userId: string;

  constructor(clientId: string, clientSecret: string, refreshToken: string, userId?: string) {
    const auth = new OAuth2Client({
      clientId: clientId,
      clientSecret: clientSecret,
      credentials: {
        refresh_token: refreshToken
      }
    });
    this.gmail = google.gmail({ auth, version: 'v1' });
    this.userId = userId ? userId : 'me';
  }

  async getEmails(params?: SearchParams): Promise<Email[]> {
    const queryParts: string[] = [];

    if (params?.unread) queryParts.push('is:unread');
    if (params?.to) queryParts.push(`to:${params.to}`);
    if (params?.from) queryParts.push(`from:${params.from}`);

    const listRes = await this.gmail.users.messages.list({
      userId: this.userId,
      q: queryParts.join(' ')
    });

    const messageIds = listRes.data.messages ?? [];

    const emails = await Promise.all(messageIds.map(msg => this.fetchMessage(msg.id!)));

    return emails;
  }

  private extractBody(payload?: gmail_v1.Schema$MessagePart): string {
    if (payload?.body?.data) {
      return Buffer.from(payload.body.data, 'base64url').toString('utf-8');
    }

    const textPart = payload?.parts?.find(p => p.mimeType === 'text/plain');

    if (textPart?.body?.data) {
      return Buffer.from(textPart.body.data, 'base64url').toString('utf-8');
    }

    return '';
  }

  private async fetchMessage(id: string): Promise<Email> {
    const res = await this.gmail.users.messages.get({
      userId: 'me',
      id,
      format: 'full'
    });

    const headers = res.data.payload?.headers ?? [];
    const header = (name: string) => headers.find(h => h.name?.toLowerCase() === name.toLowerCase())?.value ?? '';

    return {
      id,
      from: header('From'),
      to: header('To')
        .split(',')
        .map(addr => addr.trim()),
      subject: header('Subject'),
      body: this.extractBody(res.data.payload),
      receivedAt: new Date(Number(res.data.internalDate))
    };
  }
}

interface Smtp4devConfig {
  baseUrl: string; // e.g. "http://localhost:8025"
  mailboxName?: string; // defaults to "Default"
  pageSize?: number; // defaults to 50
}

interface Smtp4devMessageSummary {
  attachmentCount: number;
  deliveredTo: string | null;
  from: string | null;
  hasWarnings: boolean;
  id: string;
  isRelayed: boolean;
  isUnread: boolean;
  receivedDate: string;
  subject: string | null;
  to: string[] | null;
}

interface Smtp4devPagedResult {
  currentPage: number;
  pageCount: number;
  pageSize: number;
  results: Smtp4devMessageSummary[] | null;
  rowCount: number;
}

export class Smtp4devEmailClient implements EmailClient {
  private baseUrl: string;
  private mailboxName: string;
  private pageSize: number;

  constructor(config: Smtp4devConfig) {
    this.baseUrl = config.baseUrl.replace(/\/$/, '');
    this.mailboxName = config.mailboxName ?? 'Default';
    this.pageSize = config.pageSize ?? 50;
  }

  async getEmails(params?: SearchParams): Promise<Email[]> {
    // Use searchTerms for a rough server-side filter, then narrow client-side.
    // searchTerms searches across from, to, cc, subject, body, and attachments.
    const searchTerms = params?.from ?? params?.to ?? undefined;

    const allUnread: Email[] = [];
    let page = 1;
    let pageCount = 1;

    while (page <= pageCount) {
      const url = new URL(`${this.baseUrl}/api/Messages`);
      url.searchParams.set('mailboxName', this.mailboxName);
      url.searchParams.set('folderName', 'INBOX');
      url.searchParams.set('sortColumn', 'receivedDate');
      url.searchParams.set('sortIsDescending', 'true');
      url.searchParams.set('page', String(page));
      url.searchParams.set('pageSize', String(this.pageSize));
      if (searchTerms) url.searchParams.set('searchTerms', searchTerms);

      const res = await fetch(url.toString());
      if (!res.ok) {
        throw new Error(`smtp4dev API error: ${res.status} ${res.statusText}`);
      }

      const data: Smtp4devPagedResult = await res.json();
      pageCount = data.pageCount;

      const matches = (data.results ?? []).filter(msg => {
        if (params?.unread && !msg.isUnread) return false;
        if (params?.from && msg.from !== params.from) return false;
        if (params?.to && !msg.to?.includes(params.to)) return false;
        return true;
      });

      const emails = await Promise.all(matches.map(msg => this.hydrate(msg)));
      allUnread.push(...emails);
      page++;
    }

    return allUnread;
  }

  private async fetchPlainText(id: string): Promise<string> {
    const res = await fetch(`${this.baseUrl}/api/Messages/${id}/plaintext`);
    if (!res.ok) return '';
    return res.text();
  }

  private async hydrate(summary: Smtp4devMessageSummary): Promise<Email> {
    const body = await this.fetchPlainText(summary.id);

    return {
      id: summary.id,
      from: summary.from ?? '',
      to: summary.to ?? [],
      subject: summary.subject ?? '',
      body,
      receivedAt: new Date(summary.receivedDate)
    };
  }
}
