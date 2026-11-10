import 'dart:math' as math;

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:phone/core/constants/app_radius.dart';
import 'package:phone/features/students/extensions/student_balance_extension.dart';
import 'package:phone/features/students/extensions/student_status_extension.dart';
import 'package:phone/features/students/widgets/student_hero_card.dart';
import 'package:phone/generated/models/student_status.dart';
import 'package:phone/generated/models/student_summary_response.dart';
import 'package:phone/generated/models/transaction_response.dart';
import 'package:skeletonizer/skeletonizer.dart';

class StudentHeaderDelegate extends SliverPersistentHeaderDelegate {
  static const Color expandedColor = Color(0xFFCABEE4);
  static const double contentHeight = 136.0;

  final StudentSummaryResponse student;
  final List<TransactionResponse> payments;
  final double topPadding;
  final bool isRefreshing;
  final Widget? menu;
  final DateTime? now;
  final double? customContentHeight;

  const new({
    required this.student,
    required this.topPadding,
    this.payments = const [],
    this.isRefreshing = false,
    this.menu,
    this.now,
    this.customContentHeight,
  });

  static double computeContentHeight(
    StudentSummaryResponse student, {
    List<TransactionResponse> payments = const [],
    DateTime? now,
  }) {
    // Card vertical padding: 16 top + 16 bottom
    double cardHeight = 32.0;

    const double avatarHeight = 36.0;

    // Full name: fontSize 16, height 1.5 -> 24.0.
    double columnHeight = 24.0;

    final groups = student.studentGroups;
    if (groups != null && groups.isNotEmpty) {
      for (final group in groups) {
        if (group.name.trim().isNotEmpty) {
          columnHeight += 2.0 + 20.0;
        }
      }
    }

    if (student.birthdate != null) {
      columnHeight += 2.0 + 20.0;
    }

    if (student.phone != null && student.phone!.isNotEmpty) {
      columnHeight += 2.0 + 20.0;
    }

    columnHeight += 8.0; // SizedBox(height: 8) before badges

    int badgeCount = 0;
    final statusSet = student.statuses.toSet();
    if (statusSet.contains(StudentStatus.present) ||
        statusSet.contains(StudentStatus.missing)) {
      badgeCount++;
    }
    if (student.free ||
        statusSet.contains(StudentStatus.free) ||
        statusSet.contains(StudentStatus.paid) ||
        statusSet.contains(StudentStatus.unpaid) ||
        statusSet.contains(StudentStatus.inactive)) {
      badgeCount++;
    }
    final visibility = StudentBalanceVisibility.of(student, payments, now);
    if (visibility.showSessions) {
      badgeCount++;
    }

    if (badgeCount > 0) {
      final badgeLines = (badgeCount / 3).ceil().clamp(1, 4);
      columnHeight += badgeLines * 28.0;
    }

    final rowHeight = math.max(avatarHeight, columnHeight);
    cardHeight += rowHeight;

    final hasNote = student.note != null && student.note!.trim().isNotEmpty;
    if (hasNote) {
      cardHeight += 25.0; // 12 + 1 + 12

      final tp = TextPainter(
        text: TextSpan(
          text: student.note!,
          style: const TextStyle(fontSize: 14),
        ),
        textDirection: TextDirection.ltr,
      )..layout(maxWidth: 300);
      cardHeight += math.max(20.0, tp.height);
    }

    return cardHeight;
  }

  double get effectiveContentHeight =>
      customContentHeight ??
      computeContentHeight(student, payments: payments, now: now);

  @override
  double get minExtent => topPadding + kToolbarHeight;

  @override
  double get maxExtent => topPadding + kToolbarHeight + effectiveContentHeight;

  @override
  Widget build(
    BuildContext context,
    double shrinkOffset,
    bool overlapsContent,
  ) {
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;
    final maxShrink = maxExtent - minExtent;
    final progress = maxShrink > 0
        ? (shrinkOffset / maxShrink).clamp(0.0, 1.0)
        : 0.0;

    // Smooth cross-fade between expanded and collapsed states
    final expandedOpacity = (1.0 - progress * 1.5).clamp(0.0, 1.0);
    final collapsedOpacity = ((progress - 0.35) / 0.65).clamp(0.0, 1.0);

    final curveProgress = Curves.easeInOut.transform(progress);
    final headerColor = Color.lerp(
      expandedColor,
      colorScheme.surface,
      curveProgress,
    )!;

    final badges = [
      ...student.toHeaderStatusBadges(context),
      ...student.toBalanceBadges(context, payments, now),
    ];

    const overlayStyle = SystemUiOverlayStyle(
      statusBarColor: Colors.transparent,
      statusBarIconBrightness: Brightness.dark,
      statusBarBrightness: Brightness.light,
    );

    // Purple header banner covers status bar, app bar, and top half of the card
    final cardHeight = effectiveContentHeight;
    final bannerOffset = math.min(136.0, cardHeight * 0.55);
    final currentBannerHeight = math.max(
      minExtent,
      (topPadding + kToolbarHeight + bannerOffset) - shrinkOffset,
    );
    final bannerBorderRadius = BorderRadius.lerp(
      BorderRadius.zero,
      AppRadius.bottomRounded,
      curveProgress,
    )!;

    return AnnotatedRegion<SystemUiOverlayStyle>(
      value: overlayStyle,
      child: Stack(
        fit: StackFit.expand,
        children: [
          // 1. Purple Header Banner (covers status bar, app bar, and top half of the card)
          Positioned(
            top: 0,
            left: 0,
            right: 0,
            height: currentBannerHeight,
            child: Material(
              color: headerColor,
              shape: RoundedRectangleBorder(borderRadius: bannerBorderRadius),
              clipBehavior: Clip.antiAlias,
              elevation: 0,
              child: const SizedBox.expand(),
            ),
          ),

          // 2. Student Card (half on header, half on body)
          if (expandedOpacity > 0)
            Positioned(
              top: (topPadding + kToolbarHeight) - shrinkOffset,
              left: 16,
              right: 16,
              child: Opacity(
                opacity: expandedOpacity,
                child: IgnorePointer(
                  ignoring: expandedOpacity < 0.2,
                  child: Skeletonizer(
                    enabled: isRefreshing,
                    child: StudentHeroCard(
                      student: student,
                      overrideBadges: badges,
                    ),
                  ),
                ),
              ),
            ),

          // 3. Collapsed / Pinned App Bar
          Positioned(
            top: 0,
            left: 0,
            right: 0,
            height: minExtent,
            child: AppBar(
              systemOverlayStyle: overlayStyle,
              primary: true,
              backgroundColor: Colors.transparent,
              surfaceTintColor: Colors.transparent,
              elevation: 0,
              scrolledUnderElevation: 0,
              leading: const BackButton(),
              title: collapsedOpacity > 0
                  ? Opacity(
                      opacity: collapsedOpacity,
                      child: Text(
                        student.fullName,
                        style: const TextStyle(
                          fontSize: 20,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    )
                  : null,
              actions: [?menu, const SizedBox(width: 8)],
            ),
          ),
        ],
      ),
    );
  }

  @override
  bool shouldRebuild(covariant StudentHeaderDelegate oldDelegate) =>
      oldDelegate.student != student ||
      !listEquals(oldDelegate.payments, payments) ||
      oldDelegate.topPadding != topPadding ||
      oldDelegate.isRefreshing != isRefreshing ||
      oldDelegate.menu != menu ||
      oldDelegate.now != now ||
      oldDelegate.customContentHeight != customContentHeight;
}
