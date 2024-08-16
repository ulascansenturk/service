--
-- PostgreSQL database dump
--

-- Dumped from database version 13.3 (Debian 13.3-1.pgdg100+1)
-- Dumped by pg_dump version 16.3

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: root
--

-- *not* creating schema, since initdb creates it


ALTER SCHEMA public OWNER TO root;

--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: accounts; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE public.accounts (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    balance numeric(15,2) NOT NULL,
    currency character varying(3) NOT NULL,
    status character varying(50) NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT accounts_balance_check CHECK ((balance >= (0)::numeric))
);


ALTER TABLE public.accounts OWNER TO root;

--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


ALTER TABLE public.schema_migrations OWNER TO root;

--
-- Name: transactions; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE public.transactions (
    id uuid NOT NULL,
    user_id uuid,
    amount integer NOT NULL,
    account_id uuid NOT NULL,
    currency_code character varying(10) NOT NULL,
    reference_id uuid NOT NULL,
    metadata jsonb,
    status character varying(20) NOT NULL,
    transaction_type character varying(20) NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.transactions OWNER TO root;

--
-- Name: users; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE public.users (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    email character varying(255) NOT NULL,
    password_hash character varying(255) NOT NULL,
    first_name character varying(100) NOT NULL,
    last_name character varying(100) NOT NULL,
    date_of_birth date NOT NULL,
    phone_number character varying(20),
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.users OWNER TO root;

--
-- Name: accounts accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT accounts_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: transactions transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_pkey PRIMARY KEY (id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_accounts_currency; Type: INDEX; Schema: public; Owner: root
--

CREATE INDEX idx_accounts_currency ON public.accounts USING btree (currency);


--
-- Name: idx_accounts_status; Type: INDEX; Schema: public; Owner: root
--

CREATE INDEX idx_accounts_status ON public.accounts USING btree (status);


--
-- Name: idx_accounts_user_id; Type: INDEX; Schema: public; Owner: root
--

CREATE INDEX idx_accounts_user_id ON public.accounts USING btree (user_id);


--
-- Name: idx_transactions_account_id; Type: INDEX; Schema: public; Owner: root
--

CREATE INDEX idx_transactions_account_id ON public.transactions USING btree (account_id);


--
-- Name: idx_transactions_account_id_status; Type: INDEX; Schema: public; Owner: root
--

CREATE INDEX idx_transactions_account_id_status ON public.transactions USING btree (account_id, status);


--
-- Name: idx_transactions_account_id_type; Type: INDEX; Schema: public; Owner: root
--

CREATE INDEX idx_transactions_account_id_type ON public.transactions USING btree (account_id, transaction_type);


--
-- Name: idx_transactions_currency_code; Type: INDEX; Schema: public; Owner: root
--

CREATE INDEX idx_transactions_currency_code ON public.transactions USING btree (currency_code);


--
-- Name: idx_transactions_reference_id; Type: INDEX; Schema: public; Owner: root
--

CREATE INDEX idx_transactions_reference_id ON public.transactions USING btree (reference_id);


--
-- Name: idx_transactions_status; Type: INDEX; Schema: public; Owner: root
--

CREATE INDEX idx_transactions_status ON public.transactions USING btree (status);


--
-- Name: idx_transactions_transaction_type; Type: INDEX; Schema: public; Owner: root
--

CREATE INDEX idx_transactions_transaction_type ON public.transactions USING btree (transaction_type);


--
-- Name: idx_transactions_user_id; Type: INDEX; Schema: public; Owner: root
--

CREATE INDEX idx_transactions_user_id ON public.transactions USING btree (user_id);


--
-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: root
--

CREATE INDEX idx_users_email ON public.users USING btree (email);


--
-- Name: idx_users_id; Type: INDEX; Schema: public; Owner: root
--

CREATE INDEX idx_users_id ON public.users USING btree (id);


--
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: root
--

REVOKE USAGE ON SCHEMA public FROM PUBLIC;
GRANT ALL ON SCHEMA public TO PUBLIC;


--
-- PostgreSQL database dump complete
--

