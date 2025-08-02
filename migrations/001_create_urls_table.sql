-- 001_create_urls_table.sql
-- URL 단축 서비스를 위한 기본 테이블 생성

-- URLs 테이블
CREATE TABLE IF NOT EXISTS urls (
    id VARCHAR(255) PRIMARY KEY,
    original_url TEXT NOT NULL,
    description TEXT,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    click_count BIGINT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_accessed_at TIMESTAMP WITH TIME ZONE,
    created_by_api_key VARCHAR(255) NOT NULL
);

-- 인덱스 생성
CREATE INDEX IF NOT EXISTS idx_urls_created_by_api_key ON urls(created_by_api_key);
CREATE INDEX IF NOT EXISTS idx_urls_created_at ON urls(created_at);
CREATE INDEX IF NOT EXISTS idx_urls_expires_at ON urls(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_urls_is_active ON urls(is_active);
CREATE INDEX IF NOT EXISTS idx_urls_click_count ON urls(click_count);

-- 클릭 이벤트 테이블 (분석용)
CREATE TABLE IF NOT EXISTS click_events (
    id BIGSERIAL PRIMARY KEY,
    url_id VARCHAR(255) NOT NULL REFERENCES urls(id) ON DELETE CASCADE,
    ip_address INET NOT NULL,
    user_agent TEXT,
    referer TEXT,
    country VARCHAR(100),
    city VARCHAR(200),
    browser VARCHAR(100),
    os VARCHAR(100),
    device VARCHAR(100),
    clicked_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 클릭 이벤트 인덱스
CREATE INDEX IF NOT EXISTS idx_click_events_url_id ON click_events(url_id);
CREATE INDEX IF NOT EXISTS idx_click_events_clicked_at ON click_events(clicked_at);
CREATE INDEX IF NOT EXISTS idx_click_events_ip_address ON click_events(ip_address);
CREATE INDEX IF NOT EXISTS idx_click_events_country ON click_events(country);

-- 파티셔닝을 위한 준비 (클릭 이벤트 테이블을 월별로 파티션)
-- 실제 프로덕션에서는 월별 파티션 테이블을 생성하여 성능 최적화

-- updated_at 자동 업데이트를 위한 트리거 함수
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- updated_at 트리거 생성
CREATE TRIGGER update_urls_updated_at 
    BEFORE UPDATE ON urls 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- 만료된 URL 정리를 위한 함수
CREATE OR REPLACE FUNCTION cleanup_expired_urls()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    UPDATE urls 
    SET is_active = false, updated_at = NOW()
    WHERE expires_at < NOW() AND is_active = true;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- 오래된 클릭 이벤트 삭제 함수 (6개월 이상 된 데이터)
CREATE OR REPLACE FUNCTION cleanup_old_click_events()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM click_events 
    WHERE clicked_at < NOW() - INTERVAL '6 months';
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;